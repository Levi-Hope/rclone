package baidupan

// TODO: Solve the hotfix in lib/oauthutil/oauthutil.go:configExchange()
// TODO: Add support for refreshing Access_token: https://pan.baidu.com/union/document/entrance#3%E8%8E%B7%E5%8F%96%E6%8E%88%E6%9D%83

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/obscure"
	"github.com/rclone/rclone/fs/fshttp"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/lib/oauthutil"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	rcloneClientID              = "fP5NRUFeA3GfZpc7LuRLRTsWGSm93lmk"
	rcloneEncryptedClientSecret = "Q2m4aEy7oRoTyRe0UhWZ5ZSBrqistZAX"
)

type Fs struct {
	name     string                 // name of this remote
	root     string                 // the path we are working on
	opt      Options                // parsed options
	ci       *fs.ConfigInfo         // global config
	features *fs.Features           // optional features
	ts       *oauthutil.TokenSource // token source for oauth2
	m        configmap.Mapper
	////srv          *rest.Client       // the connection to the one drive server
	//dirCache     *dircache.DirCache // Map of directory path to directory id
	//pacer        *fs.Pacer          // pacer for API calls
	//tokenRenewer *oauthutil.Renew   // renew the token on expiry
}

func (f Fs) Name() string {
	return f.name
}

func (f Fs) Root() string {
	return f.root
}

func (f Fs) String() string {
	return fmt.Sprintf("Baidu Pan root '%s'", f.root)
}

func (f Fs) Precision() time.Duration {
	return time.Second
}

func (f Fs) Hashes() hash.Set {
	return hash.Set(hash.MD5)
}

func (f Fs) Features() *fs.Features {
	return f.features
}

type BaidupanThumbs struct {
	Url1 string `json:"url1"`
	Url2 string `json:"url2"`
	Url3 string `json:"url3"`
}

type BaidupanList struct {
	Category        int64            `json:"category"`
	Fs_id           int64            `json:"fs_id"`
	Isdir           int64            `json:"isdir"`
	Local_ctime     int64            `json:"local_ctime"`
	Local_mtime     int64            `json:"local_mtime"`
	Md5             string           `json:"md5"`
	Path            string           `json:"path"`
	Server_ctime    int64            `json:"server_ctime"`
	Server_filename string           `json:"server_filename"`
	Server_mtime    int64            `json:"server_mtime"`
	Size            int64            `json:"size"`
	Thumbs          []BaidupanThumbs `json:"thumbs"`
}

type BaidupanAPIResponse struct {
	Cursor   int64          `json:"cursor"`
	Errno    int64          `json:"errno"`
	Errmsg   string         `json:"errmsg"`
	Has_more int64          `json:"has_more"`
	List     []BaidupanList `json:"list"`
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
func (f Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {

	fmt.Println("dir: ", dir)
	fmt.Println("f.root: ", f.root)

	token, err := f.ts.Token()
	url := fmt.Sprintf("http://pan.baidu.com/rest/2.0/xpan/multimedia?method=listall&path=/%s&access_token=%s&web=1&recursion=1&start=0&limit=50", f.root, token.AccessToken)
	fs.Debugf(f, "Getting url: %s", url)

	// TODO: Use rclone API instead of http.Get
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	s := new(BaidupanAPIResponse)
	err = json.Unmarshal(body, &s)
	// TODO: Why cant unmarshal BaidupanThumbs object
	if err.Error() == "json: cannot unmarshal object into Go struct field BaidupanList.list.thumbs of type []baidupan.BaidupanThumbs" {
		fs.Debugf(f, "FIXME: %s", err)
	} else if err != nil {
		return nil, err
	}

	fmt.Println(s.Cursor, s.Errno, s.Errmsg, s.Has_more)
	for i, file := range s.List {
		fmt.Println(i, file)
	}

	// TODO: Return entries
	return nil, nil
}

func (f Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	panic("implement me")
}

func (f Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	panic("implement me")
}

func (f Fs) Mkdir(ctx context.Context, dir string) error {
	panic("implement me")
}

func (f Fs) Rmdir(ctx context.Context, dir string) error {
	panic("implement me")
}

type Options struct {
	tv oauth2.AuthCodeOption
}

type Features struct {
}

var commandHelp = []fs.CommandHelp{{}}

var Endpoint = oauth2.Endpoint{
	AuthURL:   "http://openapi.baidu.com/oauth/2.0/authorize",
	TokenURL:  "http://openapi.baidu.com/oauth/2.0/token",
	AuthStyle: oauth2.AuthStyleInHeader,
}

var (
	oauthConfig = &oauth2.Config{
		Scopes:       []string{"basic,netdisk"},
		Endpoint:     Endpoint,
		ClientID:     rcloneClientID,
		ClientSecret: obscure.MustReveal(rcloneEncryptedClientSecret),
		RedirectURL:  oauthutil.RedirectLocalhostURL,
	}
)

func Config(ctx context.Context, name string, m configmap.Mapper, config fs.ConfigIn) (*fs.ConfigOut, error) {

	if config.State == "" {

		// See: https://pan.baidu.com/union/document/entrance#3%E8%8E%B7%E5%8F%96%E6%8E%88%E6%9D%83
		opts := []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("display", "tv"),
			oauth2.SetAuthURLParam("qrcode", "1"),
			oauth2.SetAuthURLParam("force_login", "1"),
		}

		return oauthutil.ConfigOut("choose_type", &oauthutil.Options{
			OAuth2Config: oauthConfig,
			OAuth2Opts:   opts,
		})
	}

	switch config.State {
	case "choose_type":
		return fs.ConfigGoto(config.Result)
	}
	return nil, fmt.Errorf("unknown state %q", config.State)
}

func init() {
	fmt.Println("baidupan init")
	fs.Register(&fs.RegInfo{
		Name:        "baidupan",
		Description: "Baidu Wangpan",
		NewFs:       NewFs,
		CommandHelp: commandHelp,
		Config:      Config,
		//Options: append(oauthutil.SharedOptions),
	})

}

func NewFs(ctx context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {

	opt := new(Options)
	ci := fs.GetConfig(ctx)

	baseClient := fshttp.NewClient(ctx)
	_, ts, err := oauthutil.NewClientWithBaseClient(ctx, name, m, oauthConfig, baseClient)
	//fmt.Println("ts: ", ts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to configure Box")
	}

	f := &Fs{
		name: name,
		root: root,
		opt:  *opt,
		ci:   ci,
		ts:   ts,
		m:    m,
	}
	f.features = (&fs.Features{}).Fill(ctx, f)

	return f, nil
}
