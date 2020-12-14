package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Record struct {
	Advertiser   string
	AccountID    string
	Relationship string
	AuthorityID  string
}

func (r *Record) UniqueID() string {
	return r.Advertiser + r.AccountID + r.Relationship
}

func (r *Record) Row() string {
	c := []string{
		r.Advertiser,
		r.AccountID,
		r.Relationship,
	}

	if r.AuthorityID != "" {
		c = append(c, r.AuthorityID)
	}

	return strings.Join(c, ",")
}

func ParseRow(row string) (*Record, error) {
	row = strings.TrimSpace(row)
	re := regexp.MustCompile("\\s*#.+$")
	cols := strings.Split(re.ReplaceAllString(row, ""), ",")

	if len(cols) < 3 || len(cols) > 4 {
		return nil, fmt.Errorf("Failed parsing line")
	}

	r := &Record{
		Advertiser:   strings.TrimSpace(strings.ToLower(cols[0])),
		AccountID:    strings.TrimSpace(cols[1]),
		Relationship: strings.TrimSpace(cols[2]),
	}

	if len(cols) == 4 {
		r.AuthorityID = strings.TrimSpace(cols[3])
	}

	return r, nil
}

func main() {
	// load authority ids
	authorities := make(map[string]string)

	authf, err := os.Open("authorities.csv")

	if err != nil {
		log.Fatal(err)
	}

	defer authf.Close()

	auth := csv.NewReader(authf)

	for {
		record, err := auth.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		authorities[record[0]] = record[1]
	}

	// prepare file header
	header := []string{
		fmt.Sprintf("# ads.txt:%s", time.Now().Format(time.RFC3339)),
		"contact=sales@wasdmedia.de",
		"contact=https://wasdmedia.de/",
		"# ---",
	}

	// load and parse ads.txt parts
	dir := "./parts"
	rows := make(map[string]*Record)
	log.SetOutput(os.Stderr)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		f, err := os.Open(filepath.Join(dir, file.Name()))

		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			row := strings.TrimSpace(scanner.Text())

			if row[0] == '#' {
				if strings.HasPrefix(row, "#dailymotion") {
					header = append(header, "# dailymotion ads.txt version:")
					header = append(header, row)
				}

				continue
			}

			r, err := ParseRow(row)

			if err != nil {
				log.Printf("%s: %s\n", err, scanner.Text())
				continue
			}

			if id, ok := authorities[r.Advertiser]; ok {
				r.AuthorityID = id
			}

			rows[r.UniqueID()] = r
		}

		f.Close()
	}

	out := []string{}

	for _, r := range rows {
		out = append(out, r.Row())
	}

	sort.Strings(out)

	for _, h := range header {
		fmt.Println(h)
	}

	fmt.Println("# ---")

	for _, o := range out {
		fmt.Println(o)
	}
}
