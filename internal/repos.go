package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hydrogen18/stalecucumber"
	"github.com/knqyf263/berkeleydb"
	"github.com/pkg/errors"
)

const (
	// DatabaseDir is the folder where the ovs-agent databases
	// are stored.
	DatabaseDir = "/etc/ovs-agent/db"

	// RepositoryDB is the name of the repository database
	// inside DatabaseDir.
	RepositoryDB = "repository"
)

// Repo holds information about a single repository
type Repo struct {
	ID          string
	MountPoint  string `pickle:"mount_point"`
	ManagerUUID string `pickle:"manager_uuid"`
	Alias       string `pickle:"alias"`
	Version     string `pickle:"version"`
	Filesystem  string `pickle:"filesystem"`
	FSLocation  string `pickle:"fs_location"`
}

// RepoMetaItem holds one item of repository metadata
type RepoMetaItem struct {
	ObjectType string `json:"OBJECT_TYPE"`
	SimpleName string `json:"SIMPLE_NAME"`
}

// Meta returns the repository metadata
func (r *Repo) Meta() (map[string]RepoMetaItem, error) {
	repoMeta := map[string]RepoMetaItem{}

	metaFile := filepath.Join(r.MountPoint, ".ovsmeta")

	if _, err := os.Stat(metaFile); err != nil {
		return repoMeta, errors.Wrap(err, "looking for ovsmeta")
	}

	data, err := ioutil.ReadFile(metaFile)
	if err != nil {
		return repoMeta, errors.Wrap(err, "reading ovsmeta")
	}

	if err := json.Unmarshal(data, &repoMeta); err != nil {
		return repoMeta, errors.Wrap(err, "unmarshaling ovsmeta")
	}
	return repoMeta, nil
}

// ParseRepos returns a list of repositories configured
// on this compute node.
func ParseRepos() ([]Repo, error) {
	var err error

	repoFile := filepath.Join(DatabaseDir, RepositoryDB)
	if _, err := os.Stat(repoFile); err != nil {
		return nil, errors.Wrap(err, "looking up repo file")
	}

	db, err := berkeleydb.NewDB()
	if err != nil {
		return nil, fmt.Errorf("unexpected failure of CreateDB %s", err)
	}

	err = db.Open(repoFile, berkeleydb.DbHash, berkeleydb.DbRdOnly)
	if err != nil {
		return nil, fmt.Errorf("Could not open test_db.db. Error code %s", err)

	}
	defer db.Close()

	cursor, err := db.Cursor()
	if err != nil {
		return nil, fmt.Errorf("failed to create cursor: %s", err)
	}

	var ret []Repo

	for {
		k, v, err := cursor.GetNext()
		if err != nil {
			break
		}

		var repoInfo Repo
		reader := bytes.NewReader(v)
		err = stalecucumber.UnpackInto(&repoInfo).From(stalecucumber.Unpickle(reader))
		if err != nil {
			return ret, errors.Wrap(err, "decoding repo")
		}

		repoInfo.ID = string(k)
		ret = append(ret, repoInfo)
	}

	return ret, nil
}

// GetRepo returns a repo identified by repoID
func GetRepo(repoID string) (Repo, error) {
	repos, err := ParseRepos()
	if err != nil {
		return Repo{}, errors.Wrap(err, "getting repo")
	}

	for _, val := range repos {
		if val.ID == repoID {
			return val, nil
		}
	}

	return Repo{}, fmt.Errorf("could not find repo %s", repoID)
}
