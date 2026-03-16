package database

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"fpreg/internal/models"
	"fpreg/internal/repository"

	"gorm.io/gorm"
)

// LoadFacilitiesFromFile reads facilities_private.csv from the given path (or "facilities_private.csv" in cwd)
// and upserts into the facilities table by uid.
// CSV columns: uid,name,code,level,subcounty,hsd,district,client_code_prefix
// Header row is required. uid is used to uniquely identify facilities.
func LoadFacilitiesFromFile(db *gorm.DB, path string) (loaded int, err error) {
	if path == "" {
		path = "facilities_private.csv"
	}
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	repo := repository.NewFacilityRepository(db)
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}
	if len(rows) < 2 {
		log.Println("LoadFacilities: no data rows in", path)
		return 0, nil
	}

	header := rows[0]
	uidIdx, nameIdx, codeIdx, levelIdx, subIdx, hsdIdx, distIdx, prefixIdx := -1, -1, -1, -1, -1, -1, -1, -1
	for i, h := range header {
		switch strings.TrimSpace(strings.ToLower(h)) {
		case "uid":
			uidIdx = i
		case "name":
			nameIdx = i
		case "code":
			codeIdx = i
		case "level":
			levelIdx = i
		case "subcounty":
			subIdx = i
		case "hsd":
			hsdIdx = i
		case "district":
			distIdx = i
		case "client_code_prefix", "client_code_prefix ", "prefix":
			prefixIdx = i
		}
	}
	if uidIdx < 0 || nameIdx < 0 || codeIdx < 0 || prefixIdx < 0 {
		log.Printf("LoadFacilities: required columns uid, name, code, client_code_prefix not found (uid=%d name=%d code=%d prefix=%d)", uidIdx, nameIdx, codeIdx, prefixIdx)
		return 0, nil
	}

	for _, row := range rows[1:] {
		if len(row) <= max(uidIdx, nameIdx, codeIdx, prefixIdx) {
			continue
		}
		uid := strings.TrimSpace(row[uidIdx])
		name := strings.TrimSpace(row[nameIdx])
		code := strings.TrimSpace(row[codeIdx])
		if uid == "" || name == "" || code == "" {
			continue
		}
		prefix := strings.TrimSpace(row[prefixIdx])
		if prefix == "" {
			prefix = code
		}
		level, sub, hsd, dist := "", "", "", ""
		if levelIdx >= 0 && levelIdx < len(row) {
			level = strings.TrimSpace(row[levelIdx])
		}
		if subIdx >= 0 && subIdx < len(row) {
			sub = strings.TrimSpace(row[subIdx])
		}
		if hsdIdx >= 0 && hsdIdx < len(row) {
			hsd = strings.TrimSpace(row[hsdIdx])
		}
		if distIdx >= 0 && distIdx < len(row) {
			dist = strings.TrimSpace(row[distIdx])
		}

		fac := &models.Facility{
			UID:              uid,
			Name:             name,
			Code:             code,
			Level:            level,
			Subcounty:        sub,
			HSD:              hsd,
			District:         dist,
			ClientCodePrefix: prefix,
		}
		if err := repo.UpsertByUID(fac); err != nil {
			log.Printf("LoadFacilities: upsert uid=%s: %v", uid, err)
			continue
		}
		loaded++
	}
	return loaded, nil
}

func max(a, b, c, d int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	if d > a {
		a = d
	}
	return a
}
