package union

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rclone/rclone/backend/union/policy"
	"github.com/rclone/rclone/backend/union/upstream"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/configstruct"
	"github.com/rclone/rclone/fs/hash"
)

// Register with Fs
func init() {
	fsi := &fs.RegInfo{
		Name:        "union",
		Description: "Union merges the contents of several upstream fs",
		NewFs:       NewFs,
		Options: []fs.Option{{
			Name:     "upstreams",
			Help:     "List of space separated upstreams.\nCan be 'upstreama:test/dir upstreamb:', '\"upstreama:test/space:ro dir\" upstreamb:', etc.\n",
			Required: true,
		}, {
			Name:     "action_policy",
			Help:     "Policy to choose upstream on ACTION category.",
			Required: true,
			Default:  "epall",
		}, {
			Name:     "create_policy",
			Help:     "Policy to choose upstream on CREATE category.",
			Required: true,
			Default:  "epmfs",
		}, {
			Name:     "search_policy",
			Help:     "Policy to choose upstream on SEARCH category.",
			Required: true,
			Default:  "ff",
		}, {
			Name:     "cache_time",
			Help:     "Cache time of usage and free space (in seconds). This option is only useful when a path preserving policy is used.",
			Required: true,
			Default:  120,
		}},
	}
	fs.Register(fsi)
}

// Options defines the configuration for this backend
type Options struct {
	Upstreams    fs.SpaceSepList `config:"upstreams"`
	Remotes      fs.SpaceSepList `config:"remotes"` // Depreated
	ActionPolicy string          `config:"action_policy"`
	CreatePolicy string          `config:"create_policy"`
	SearchPolicy string          `config:"search_policy"`
	CacheTime    int             `config:"cache_time"`
}

// Fs represents a union of upstreams
type Fs struct {
	name         string         // name of this remote
	features     *fs.Features   // optional features
	opt          Options        // options for this Fs
	root         string         // the path we are working on
	upstreams    []*upstream.Fs // slice of upstreams
	hashSet      hash.Set       // intersection of hash types
	actionPolicy policy.Policy  // policy for ACTION
	createPolicy policy.Policy  // policy for CREATE
	searchPolicy policy.Policy  // policy for SEARCH
}

// Wrap candidate objects in to an union Object
func (f *Fs) wrapEntries(entries ...upstream.Entry) (entry, error) {
	e, err := f.searchEntries(entries...)
	if err != nil {
		return nil, err
	}
	switch e.(type) {
	case *upstream.Object:
		return &Object{
			Object: e.(*upstream.Object),
			fs:     f,
			co:     entries,
		}, nil
	case *upstream.Directory:
		return &Directory{
			Directory: e.(*upstream.Directory),
			cd:        entries,
		}, nil
	default:
		return nil, errors.Errorf("unknown object type %T", e)
	}
}

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String converts this Fs to a string
func (f *Fs) String() string {
	return fmt.Sprintf("union root '%s'", f.root)
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// Rmdir removes the root directory of the Fs object
func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	upstreams, err := f.action(ctx, dir)
	if err != nil {
		return err
	}
	errs := Errors(make([]error, len(upstreams)))
	multithread(len(upstreams), func(i int) {
		err := upstreams[i].Rmdir(ctx, dir)
		errs[i] = errors.Wrap(err, upstreams[i].Name())
	})
	return errs.Err()
}

// Hashes returns hash.HashNone to indicate remote hashing is unavailable
func (f *Fs) Hashes() hash.Set {
	return f.hashSet
}

// Mkdir makes the root directory of the Fs object
func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	upstreams, err := f.create(ctx, dir)
	if err == fs.ErrorObjectNotFound && dir != parentDir(dir) {
		if err := f.Mkdir(ctx, parentDir(dir)); err != nil {
			return err
		}
		upstreams, err = f.create(ctx, dir)
	}
	if err != nil {
		return err
	}
	errs := Errors(make([]error, len(upstreams)))
	multithread(len(upstreams), func(i int) {
		err := upstreams[i].Mkdir(ctx, dir)
		errs[i] = errors.Wrap(err, upstreams[i].Name())
	})
	return errs.Err()
}

// Purge all files in the root and the root directory
//
// Implement this if you have a way of deleting all the files
// quicker than just running Remove() on the result of List()
//
// Return an error if it doesn't exist
func (f *Fs) Purge(ctx context.Context) error {
	for _, r := range f.upstreams {
		if r.Features().Purge == nil {
			return fs.ErrorCantPurge
		}
	}
	upstreams, err := f.action(ctx, "")
	if err != nil {
		return err
	}
	errs := Errors(make([]error, len(upstreams)))
	multithread(len(upstreams), func(i int) {
		err := upstreams[i].Features().Purge(ctx)
		errs[i] = errors.Wrap(err, upstreams[i].Name())
	})
	return errs.Err()
}

// Copy src to this remote using server side copy operations.
//
// This is stored with the remote path given
//
// It returns the destination Object and a possible error
//
// Will only be called if src.Fs().Name() == f.Name()
//
// If it isn't possible then return fs.ErrorCantCopy
func (f *Fs) Copy(ctx context.Context, src fs.Object, remote string) (fs.Object, error) {
	srcObj, ok := src.(*Object)
	if !ok {
		fs.Debugf(src, "Can't copy - not same remote type")
		return nil, fs.ErrorCantCopy
	}
	o := srcObj.UnWrap()
	u := o.UpstreamFs()
	do := u.Features().Copy
	if do == nil {
		return nil, fs.ErrorCantCopy
	}
	if !u.IsCreatable() {
		return nil, fs.ErrorPermissionDenied
	}
	co, err := do(ctx, o, remote)
	if err != nil || co == nil {
		return nil, err
	}
	wo, err := f.wrapEntries(u.WrapObject(co))
	return wo.(*Object), err
}

// Move src to this remote using server side move operations.
//
// This is stored with the remote path given
//
// It returns the destination Object and a possible error
//
// Will only be called if src.Fs().Name() == f.Name()
//
// If it isn't possible then return fs.ErrorCantMove
func (f *Fs) Move(ctx context.Context, src fs.Object, remote string) (fs.Object, error) {
	o, ok := src.(*Object)
	if !ok {
		fs.Debugf(src, "Can't move - not same remote type")
		return nil, fs.ErrorCantMove
	}
	entries, err := f.actionEntries(o.candidates()...)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.UpstreamFs().Features().Move == nil {
			return nil, fs.ErrorCantMove
		}
	}
	objs := make([]*upstream.Object, len(entries))
	errs := Errors(make([]error, len(entries)))
	multithread(len(entries), func(i int) {
		u := entries[i].UpstreamFs()
		o, ok := entries[i].(*upstream.Object)
		if !ok {
			errs[i] = errors.Wrap(fs.ErrorNotAFile, u.Name())
			return
		}
		mo, err := u.Features().Move(ctx, o.UnWrap(), remote)
		if err != nil || mo == nil {
			errs[i] = errors.Wrap(err, u.Name())
			return
		}
		objs[i] = u.WrapObject(mo)
	})
	var en []upstream.Entry
	for _, o := range objs {
		if o != nil {
			en = append(en, o)
		}
	}
	e, err := f.wrapEntries(en...)
	if err != nil {
		return nil, err
	}
	return e.(*Object), errs.Err()
}

// DirMove moves src, srcRemote to this remote at dstRemote
// using server side move operations.
//
// Will only be called if src.Fs().Name() == f.Name()
//
// If it isn't possible then return fs.ErrorCantDirMove
//
// If destination exists then return fs.ErrorDirExists
func (f *Fs) DirMove(ctx context.Context, src fs.Fs, srcRemote, dstRemote string) error {
	sfs, ok := src.(*Fs)
	if !ok {
		fs.Debugf(src, "Can't move directory - not same remote type")
		return fs.ErrorCantDirMove
	}
	upstreams, err := sfs.action(ctx, srcRemote)
	if err != nil {
		return err
	}
	for _, u := range upstreams {
		if u.Features().DirMove == nil {
			return fs.ErrorCantDirMove
		}
	}
	errs := Errors(make([]error, len(upstreams)))
	multithread(len(upstreams), func(i int) {
		su := upstreams[i]
		var du *upstream.Fs
		for _, u := range f.upstreams {
			if u.RootFs.Root() == su.RootFs.Root() {
				du = u
			}
		}
		if du == nil {
			errs[i] = errors.Wrap(fs.ErrorCantDirMove, su.Name()+":"+su.Root())
			return
		}
		err := du.Features().DirMove(ctx, su.Fs, srcRemote, dstRemote)
		errs[i] = errors.Wrap(err, du.Name()+":"+du.Root())
	})
	errs = errs.FilterNil()
	if len(errs) == 0 {
		return nil
	}
	for _, e := range errs {
		if errors.Cause(e) != fs.ErrorDirExists {
			return errs
		}
	}
	return fs.ErrorDirExists
}

// ChangeNotify calls the passed function with a path
// that has had changes. If the implementation
// uses polling, it should adhere to the given interval.
// At least one value will be written to the channel,
// specifying the initial value and updated values might
// follow. A 0 Duration should pause the polling.
// The ChangeNotify implementation must empty the channel
// regularly. When the channel gets closed, the implementation
// should stop polling and release resources.
func (f *Fs) ChangeNotify(ctx context.Context, fn func(string, fs.EntryType), ch <-chan time.Duration) {
	var uChans []chan time.Duration

	for _, u := range f.upstreams {
		if ChangeNotify := u.Features().ChangeNotify; ChangeNotify != nil {
			ch := make(chan time.Duration)
			uChans = append(uChans, ch)
			ChangeNotify(ctx, fn, ch)
		}
	}

	go func() {
		for i := range ch {
			for _, c := range uChans {
				c <- i
			}
		}
		for _, c := range uChans {
			close(c)
		}
	}()
}

// DirCacheFlush resets the directory cache - used in testing
// as an optional interface
func (f *Fs) DirCacheFlush() {
	multithread(len(f.upstreams), func(i int) {
		if do := f.upstreams[i].Features().DirCacheFlush; do != nil {
			do()
		}
	})
}

func (f *Fs) put(ctx context.Context, in io.Reader, src fs.ObjectInfo, stream bool, options ...fs.OpenOption) (fs.Object, error) {
	srcPath := src.Remote()
	upstreams, err := f.create(ctx, srcPath)
	if err == fs.ErrorObjectNotFound {
		if err := f.Mkdir(ctx, parentDir(srcPath)); err != nil {
			return nil, err
		}
		upstreams, err = f.create(ctx, srcPath)
	}
	if err != nil {
		return nil, err
	}
	if len(upstreams) == 1 {
		u := upstreams[0]
		var o fs.Object
		var err error
		if stream {
			o, err = u.Features().PutStream(ctx, in, src, options...)
		} else {
			o, err = u.Put(ctx, in, src, options...)
		}
		if err != nil {
			return nil, err
		}
		e, err := f.wrapEntries(u.WrapObject(o))
		return e.(*Object), err
	}
	errs := Errors(make([]error, len(upstreams)+1))
	// Get multiple reader
	readers := make([]io.Reader, len(upstreams))
	writers := make([]io.Writer, len(upstreams))
	for i := range writers {
		r, w := io.Pipe()
		bw := bufio.NewWriter(w)
		readers[i], writers[i] = r, bw
		defer func() {
			err := w.Close()
			if err != nil {
				panic(err)
			}
		}()
	}
	go func() {
		mw := io.MultiWriter(writers...)
		es := make([]error, len(writers)+1)
		_, es[len(es)-1] = io.Copy(mw, in)
		for i, bw := range writers {
			es[i] = bw.(*bufio.Writer).Flush()
		}
		errs[len(upstreams)] = Errors(es).Err()
	}()
	// Multi-threading
	objs := make([]upstream.Entry, len(upstreams))
	multithread(len(upstreams), func(i int) {
		u := upstreams[i]
		var o fs.Object
		var err error
		if stream {
			o, err = u.Features().PutStream(ctx, readers[i], src, options...)
		} else {
			o, err = u.Put(ctx, readers[i], src, options...)
		}
		if err != nil {
			errs[i] = errors.Wrap(err, u.Name())
			return
		}
		objs[i] = u.WrapObject(o)
	})
	err = errs.Err()
	if err != nil {
		return nil, err
	}
	e, err := f.wrapEntries(objs...)
	return e.(*Object), err
}

// Put in to the remote path with the modTime given of the given size
//
// May create the object even if it returns an error - if so
// will return the object and the error, otherwise will return
// nil and the error
func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	o, err := f.NewObject(ctx, src.Remote())
	switch err {
	case nil:
		return o, o.Update(ctx, in, src, options...)
	case fs.ErrorObjectNotFound:
		return f.put(ctx, in, src, false, options...)
	default:
		return nil, err
	}
}

// PutStream uploads to the remote path with the modTime given of indeterminate size
//
// May create the object even if it returns an error - if so
// will return the object and the error, otherwise will return
// nil and the error
func (f *Fs) PutStream(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	o, err := f.NewObject(ctx, src.Remote())
	switch err {
	case nil:
		return o, o.Update(ctx, in, src, options...)
	case fs.ErrorObjectNotFound:
		return f.put(ctx, in, src, true, options...)
	default:
		return nil, err
	}
}

// About gets quota information from the Fs
func (f *Fs) About(ctx context.Context) (*fs.Usage, error) {
	usage := &fs.Usage{
		Total:   new(int64),
		Used:    new(int64),
		Trashed: new(int64),
		Other:   new(int64),
		Free:    new(int64),
		Objects: new(int64),
	}
	for _, u := range f.upstreams {
		usg, err := u.About(ctx)
		if err != nil {
			return nil, err
		}
		if usg.Total != nil && usage.Total != nil {
			*usage.Total += *usg.Total
		} else {
			usage.Total = nil
		}
		if usg.Used != nil && usage.Used != nil {
			*usage.Used += *usg.Used
		} else {
			usage.Used = nil
		}
		if usg.Trashed != nil && usage.Trashed != nil {
			*usage.Trashed += *usg.Trashed
		} else {
			usage.Trashed = nil
		}
		if usg.Other != nil && usage.Other != nil {
			*usage.Other += *usg.Other
		} else {
			usage.Other = nil
		}
		if usg.Free != nil && usage.Free != nil {
			*usage.Free += *usg.Free
		} else {
			usage.Free = nil
		}
		if usg.Objects != nil && usage.Objects != nil {
			*usage.Objects += *usg.Objects
		} else {
			usage.Objects = nil
		}
	}
	return usage, nil
}

// List the objects and directories in dir into entries.  The
// entries can be returned in any order but should be for a
// complete directory.
//
// dir should be "" to list the root, and should not have
// trailing slashes.
//
// This should return ErrDirNotFound if the directory isn't
// found.
func (f *Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	entriess := make([][]upstream.Entry, len(f.upstreams))
	errs := Errors(make([]error, len(f.upstreams)))
	multithread(len(f.upstreams), func(i int) {
		u := f.upstreams[i]
		entries, err := u.List(ctx, dir)
		if err != nil {
			errs[i] = errors.Wrap(err, u.Name())
			return
		}
		uEntries := make([]upstream.Entry, len(entries))
		for j, e := range entries {
			uEntries[j], _ = u.WrapEntry(e)
		}
		entriess[i] = uEntries
	})
	if len(errs) == len(errs.FilterNil()) {
		errs = errs.Map(func(e error) error {
			if errors.Cause(e) == fs.ErrorDirNotFound {
				return nil
			}
			return e
		})
		if len(errs) == 0 {
			return nil, fs.ErrorDirNotFound
		}
		return nil, errs.Err()
	}
	return f.mergeDirEntries(entriess)
}

// NewObject creates a new remote union file object
func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	objs := make([]*upstream.Object, len(f.upstreams))
	errs := Errors(make([]error, len(f.upstreams)))
	multithread(len(f.upstreams), func(i int) {
		u := f.upstreams[i]
		o, err := u.NewObject(ctx, remote)
		if err != nil && err != fs.ErrorObjectNotFound {
			errs[i] = errors.Wrap(err, u.Name())
			return
		}
		objs[i] = u.WrapObject(o)
	})
	var entries []upstream.Entry
	for _, o := range objs {
		if o != nil {
			entries = append(entries, o)
		}
	}
	if len(entries) == 0 {
		return nil, fs.ErrorObjectNotFound
	}
	e, err := f.wrapEntries(entries...)
	if err != nil {
		return nil, err
	}
	return e.(*Object), errs.Err()
}

// Precision is the greatest Precision of all upstreams
func (f *Fs) Precision() time.Duration {
	var greatestPrecision time.Duration
	for _, u := range f.upstreams {
		if u.Precision() > greatestPrecision {
			greatestPrecision = u.Precision()
		}
	}
	return greatestPrecision
}

func (f *Fs) action(ctx context.Context, path string) ([]*upstream.Fs, error) {
	return f.actionPolicy.Action(ctx, f.upstreams, path)
}

func (f *Fs) actionEntries(entries ...upstream.Entry) ([]upstream.Entry, error) {
	return f.actionPolicy.ActionEntries(entries...)
}

func (f *Fs) create(ctx context.Context, path string) ([]*upstream.Fs, error) {
	return f.createPolicy.Create(ctx, f.upstreams, path)
}

func (f *Fs) createEntries(entries ...upstream.Entry) ([]upstream.Entry, error) {
	return f.createPolicy.CreateEntries(entries...)
}

func (f *Fs) search(ctx context.Context, path string) (*upstream.Fs, error) {
	return f.searchPolicy.Search(ctx, f.upstreams, path)
}

func (f *Fs) searchEntries(entries ...upstream.Entry) (upstream.Entry, error) {
	return f.searchPolicy.SearchEntries(entries...)
}

func (f *Fs) mergeDirEntries(entriess [][]upstream.Entry) (fs.DirEntries, error) {
	entryMap := make(map[string]([]upstream.Entry))
	for _, en := range entriess {
		if en == nil {
			continue
		}
		for _, entry := range en {
			remote := entry.Remote()
			if f.Features().CaseInsensitive {
				remote = strings.ToLower(remote)
			}
			entryMap[remote] = append(entryMap[remote], entry)
		}
	}
	var entries fs.DirEntries
	for path := range entryMap {
		e, err := f.wrapEntries(entryMap[path]...)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// NewFs constructs an Fs from the path.
//
// The returned Fs is the actual Fs, referenced by remote in the config
func NewFs(name, root string, m configmap.Mapper) (fs.Fs, error) {
	// Parse config into Options struct
	opt := new(Options)
	err := configstruct.Set(m, opt)
	if err != nil {
		return nil, err
	}
	// Backward compatible to old config
	if len(opt.Upstreams) == 0 && len(opt.Remotes) > 0 {
		for i := 0; i < len(opt.Remotes)-1; i++ {
			opt.Remotes[i] = opt.Remotes[i] + ":ro"
		}
		opt.Upstreams = opt.Remotes
	}
	if len(opt.Upstreams) == 0 {
		return nil, errors.New("union can't point to an empty upstream - check the value of the upstreams setting")
	}
	if len(opt.Upstreams) == 1 {
		return nil, errors.New("union can't point to a single upstream - check the value of the upstreams setting")
	}
	for _, u := range opt.Upstreams {
		if strings.HasPrefix(u, name+":") {
			return nil, errors.New("can't point union remote at itself - check the value of the upstreams setting")
		}
	}

	upstreams := make([]*upstream.Fs, len(opt.Upstreams))
	errs := Errors(make([]error, len(opt.Upstreams)))
	multithread(len(opt.Upstreams), func(i int) {
		u := opt.Upstreams[i]
		upstreams[i], errs[i] = upstream.New(u, root, time.Duration(opt.CacheTime)*time.Second)
	})
	var usedUpstreams []*upstream.Fs
	var fserr error
	for i, err := range errs {
		if err != nil && err != fs.ErrorIsFile {
			return nil, err
		}
		// Only the upstreams returns ErrorIsFile would be used if any
		if err == fs.ErrorIsFile {
			usedUpstreams = append(usedUpstreams, upstreams[i])
			fserr = fs.ErrorIsFile
		}
	}
	if fserr == nil {
		usedUpstreams = upstreams
	}

	f := &Fs{
		name:      name,
		root:      root,
		opt:       *opt,
		upstreams: usedUpstreams,
	}
	f.actionPolicy, err = policy.Get(opt.ActionPolicy)
	if err != nil {
		return nil, err
	}
	f.createPolicy, err = policy.Get(opt.CreatePolicy)
	if err != nil {
		return nil, err
	}
	f.searchPolicy, err = policy.Get(opt.SearchPolicy)
	if err != nil {
		return nil, err
	}
	var features = (&fs.Features{
		CaseInsensitive:         true,
		DuplicateFiles:          false,
		ReadMimeType:            true,
		WriteMimeType:           true,
		CanHaveEmptyDirectories: true,
		BucketBased:             true,
		SetTier:                 true,
		GetTier:                 true,
	}).Fill(f)
	for _, f := range upstreams {
		if !f.IsWritable() {
			continue
		}
		features = features.Mask(f) // Mask all writable upstream fs
	}

	// Really need the union of all upstreams for these, so
	// re-instate and calculate separately.
	features.ChangeNotify = f.ChangeNotify
	features.DirCacheFlush = f.DirCacheFlush

	// FIXME maybe should be masking the bools here?

	// Clear ChangeNotify and DirCacheFlush if all are nil
	clearChangeNotify := true
	clearDirCacheFlush := true
	for _, u := range f.upstreams {
		uFeatures := u.Features()
		if uFeatures.ChangeNotify != nil {
			clearChangeNotify = false
		}
		if uFeatures.DirCacheFlush != nil {
			clearDirCacheFlush = false
		}
	}
	if clearChangeNotify {
		features.ChangeNotify = nil
	}
	if clearDirCacheFlush {
		features.DirCacheFlush = nil
	}

	f.features = features

	// Get common intersection of hashes
	hashSet := f.upstreams[0].Hashes()
	for _, u := range f.upstreams[1:] {
		hashSet = hashSet.Overlap(u.Hashes())
	}
	f.hashSet = hashSet

	return f, fserr
}

func parentDir(absPath string) string {
	parent := path.Dir(strings.TrimRight(filepath.ToSlash(absPath), "/"))
	if parent == "." {
		parent = ""
	}
	return parent
}

func multithread(num int, fn func(int)) {
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			fn(i)
		}()
	}
	wg.Wait()
}

// Check the interfaces are satisfied
var (
	_ fs.Fs              = (*Fs)(nil)
	_ fs.Purger          = (*Fs)(nil)
	_ fs.PutStreamer     = (*Fs)(nil)
	_ fs.Copier          = (*Fs)(nil)
	_ fs.Mover           = (*Fs)(nil)
	_ fs.DirMover        = (*Fs)(nil)
	_ fs.DirCacheFlusher = (*Fs)(nil)
	_ fs.ChangeNotifier  = (*Fs)(nil)
	_ fs.Abouter         = (*Fs)(nil)
)
