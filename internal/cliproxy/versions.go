package cliproxy

import (
	"context"
	"encoding/json"
)

func Versions(ctx context.Context) ([]string, error) {
	data, err := download(ctx, "https://api.github.com/repos/router-for-me/CLIProxyAPI/releases?per_page=30", 4<<20)
	if err != nil {
		return nil, err
	}
	var releases []struct {
		Tag        string `json:"tag_name"`
		Draft      bool
		Prerelease bool
	}
	if err = json.Unmarshal(data, &releases); err != nil {
		return nil, err
	}
	versions := []string{}
	for _, release := range releases {
		if !release.Draft && !release.Prerelease && ValidVersion(release.Tag) {
			versions = append(versions, release.Tag)
		}
	}
	return versions, nil
}
