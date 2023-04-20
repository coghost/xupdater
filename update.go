package xupdater

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
	"github.com/rs/zerolog/log"
)

const (
	Github = "github"
	Local  = "local"
	Zip    = "zip"
)

type XUpdater struct {
	providerType string

	URI     string
	name    string
	version string
}

// NewXUpdater creates a wrapper of rocket-update
func NewXUpdater(provider, URI, name, version string) *XUpdater {
	_os, _arch := runtime.GOOS, runtime.GOARCH
	return &XUpdater{
		providerType: provider,
		URI:          URI,
		name:         fmt.Sprintf("%s_%s_%s", name, _os, _arch),
		version:      version,
	}
}

// UpdateAndExit updates program and exit
//
//	if no error: exit(0)
//	else: exit(-1)
func (c *XUpdater) UpdateAndExit() {
	if e := c.UpdateE(); e != nil {
		log.Fatal().Err(e).Msg("update failed")
		os.Exit(-1)
	}
	os.Exit(0)
}

func (c *XUpdater) UpdateE() error {
	log.Info().Str("name", c.name).Str("url", c.URI).Str("version", c.version).Msg("try update from")

	var p provider.Provider

	switch c.providerType {
	case Github:
		// --update github,<URI>,<name>
		// and in github, create a release with name of vX.X.X and upload local zipped file
		p = c.getGithubProvider()
	case Zip:
		p = c.getZipProvider()
	default:
		// --update ,<LOCAL_DIRECTORY>,<name>
		p = c.getLocalProvider()
	}
	u := &updater.Updater{
		Provider:       p,
		ExecutableName: c.name,
		Version:        c.version,
	}

	status, err := u.Update()
	log.Info().Msgf("get response with status=%v", status)
	if err != nil {
		log.Error().Err(err).Msg("get updater failed")
		return err
	}

	switch status {
	case updater.UpToDate:
		log.Info().Msg("already up to date, no need to update!")
	case updater.Updated:
		log.Info().Msg("updated")
	case updater.Unknown:
		log.Info().Msg("something wrong happend")
		return fmt.Errorf("something wrong happend")
	}
	return nil
}

func (c *XUpdater) getGithubProvider() provider.Provider {
	return &provider.Github{
		RepositoryURL: c.URI,
		ArchiveName:   fmt.Sprintf("%s.zip", c.name),
	}
}

func (c *XUpdater) getLocalProvider() provider.Provider {
	return &provider.Local{
		Path: c.URI,
	}
}

func (c *XUpdater) getZipProvider() provider.Provider {
	fpth, e := provider.GlobNewestFile(fmt.Sprintf("%s/*.zip", c.URI))
	if e != nil {
		log.Err(e)
	}
	return &provider.Zip{
		Path: fpth,
	}
}
