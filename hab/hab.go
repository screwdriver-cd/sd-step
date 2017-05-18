package hab

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type PackagesInfo struct {
	RangeStart  int           `json:"range_start"`
	RangeEnd    int           `json:"range_end"`
	TotalCount  int           `json:"total_count"`
	PackageList []PackageInfo `json:"package_list"`
}

type PackageInfo struct {
	Origin  string `json:"origin"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Release string `json:"release"`
}

type Depot interface {
	PackageVersionsFromName(pkgName string) ([]string, error)
}

type depot struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) *depot {
	return &depot{baseURL, &http.Client{Timeout: 10 * time.Second}}
}

func (depo *depot) packagesInfo(pkgName string, from int) (PackagesInfo, error) {
	pkgUrl := fmt.Sprintf("%s/pkgs/%s?range=%d", depo.baseURL, pkgName, from)
	res, err := depo.client.Get(pkgUrl)

	if err != nil {
		return PackagesInfo{}, err
	}

	defer res.Body.Close()

	if res.StatusCode == 404 {
		return PackagesInfo{}, errors.New("Package not found")
	}

	if res.StatusCode != 200 {
		return PackagesInfo{}, errors.New(fmt.Sprintf("Unexpected status code: %d", res.StatusCode))
	}

	var pkgsInfo PackagesInfo
	json.NewDecoder(res.Body).Decode(&pkgsInfo)

	return pkgsInfo, nil
}

func (depo *depot) PackageVersionsFromName(pkgName string) ([]string, error) {
	var packages []PackageInfo

	offset := 0
	for {
		pkgsInfo, err := depo.packagesInfo(pkgName, offset)

		if err != nil {
			return nil, err
		}

		packages = append(packages, pkgsInfo.PackageList...)

		if pkgsInfo.RangeEnd+1 >= pkgsInfo.TotalCount {
			break
		}

		offset += pkgsInfo.RangeEnd - pkgsInfo.RangeStart
	}

	var versions []string
	for _, pkg := range packages {
		versions = append(versions, pkg.Version)
	}

	return versions, nil
}
