package baidupan

import (
	"context"
	"fmt"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/obscure"
	"github.com/rclone/rclone/fs/hash"
	"github.com/rclone/rclone/lib/oauthutil"
	"golang.org/x/oauth2"
	"io"
	"time"
)

const (
	// rcloneAppID = "24599545"
	rcloneClientID = "fP5NRUFeA3GfZpc7LuRLRTsWGSm93lmk"
	rcloneEncryptedClientSecret = "Q2m4aEy7oRoTyRe0UhWZ5ZSBrqistZAX"
)

type Fs struct {
	name         string             // name of this remote
	root         string             // the path we are working on
	opt          Options            // parsed options
	ci           *fs.ConfigInfo     // global config
	features     *fs.Features       // optional features
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
	return fmt.Sprintf("Baidu Pan root '%s", f.root)
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

func (f Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	panic("implement me")
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
	tv  oauth2.AuthCodeOption

}

type Features struct {

}


var commandHelp = []fs.CommandHelp{{

}}

var Endpoint = oauth2.Endpoint{
	AuthURL:   "http://openapi.baidu.com/oauth/2.0/authorize",
	TokenURL:  "http://openapi.baidu.com/oauth/2.0/token",
	//AuthStyle: oauth2.AuthStyleInParams,
}


var (
	oauthConfig = &oauth2.Config{
		Scopes:       []string{"basic,netdisk"},
		Endpoint:     Endpoint,
		ClientID:     rcloneClientID,
		ClientSecret: obscure.MustReveal(rcloneEncryptedClientSecret),
		RedirectURL:  oauthutil.RedirectLocalhostURL,
		//RedirectURL: "oob",
	}
)

func Config(ctx context.Context, name string, m configmap.Mapper, config fs.ConfigIn) (*fs.ConfigOut, error) {

	if config.State == ""{
		opts := []oauth2.AuthCodeOption{
			oauth2.SetAuthURLParam("display", "tv"),
			oauth2.SetAuthURLParam("qrcode", "1"),
			oauth2.SetAuthURLParam("force_login", "1"),
		}

		return oauthutil.ConfigOut("choose_type", &oauthutil.Options{
			OAuth2Config: oauthConfig,
			OAuth2Opts: opts,
		})
	}

	switch config.State {
	case "choose_type":
		return fs.ConfigGoto(config.Result)
	}
	return nil, fmt.Errorf("unknown state %q", config.State)
}

func init(){
	fmt.Println("baidupan init")
	fs.Register(&fs.RegInfo{
		Name:	"baidupan",
		Description: "Baidu Wangpan",
		NewFs: NewFs,
		CommandHelp: commandHelp,
		Config: Config,
		//Options: append(oauthutil.SharedOptions),

	})

}

func NewFs(ctx context.Context, name, root string, m configmap.Mapper) (fs.Fs, error) {

	opt := new(Options)
	ci := fs.GetConfig(ctx)
	f := &Fs{
		name: name,
		root: root,
		opt:  *opt,
		ci:   ci,
	}
	f.features = (&fs.Features{
	}).Fill(ctx, f)

	return f, nil
}