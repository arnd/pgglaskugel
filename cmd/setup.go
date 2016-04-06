// Copyright © 2016 Alexander Sosna <alexander@xxor.de>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	log "github.com/Sirupsen/logrus"

	// This is needed but never directly called
	_ "github.com/lib/pq"
)

// setupCmd represents the setup command
var (
	// PostgreSQL settings
	pgSettings = map[string]string{
		"archive_command": "",
		"archive_mode":    "",
		"wal_level":       "",
	}

	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup PostgreSQL and needed directories.",
		Long:  `This command makes all needed configuration changes via ALTER SYSTEM and creates missing folders. To operate it needs a superuser connection (connection sting) and the path where the backups should go.`,
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Work your own magic here
			fmt.Println("setup called")

			// Set loglevel to debug
			log.SetLevel(log.DebugLevel)

			fmt.Println("pg_gobackup")

			// Setup
			createDirs("tmp")

			// Connect to database
			conString := "user=postgres dbname=postgres password=toor"
			log.Debug("Connection string, conString:", conString)
			db, err := sql.Open("postgres", conString)
			if err != nil {
				log.Fatal("Unable to connect to database!")
			}
			defer db.Close()

			// Get version
			pgVersion, err := getPgSetting(db, "server_version")
			check(err)
			log.Debug("pgVersion ", pgVersion)

			// Configure PostgreSQL for archiving
			fmt.Println("Configure PostgreSQL for archiving.")
			changed, _ := configurePostgreSQL(db, pgSettings)
			check(err)

			if changed > 0 {
				// Configure PostgreSQL again to see if all settings are good now!
				changed, _ = configurePostgreSQL(db, pgSettings)
				check(err)
			}

			if changed > 0 {
				// Settings are still not good, restart needed!
				log.Warn("Not all settings took affect, restart the Database!")
			}

			// Create directories for backups and WAL
			createDirs("tmp/")
			check(err)
		},
	}
)

func init() {
	RootCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	pgSettings["archive_command"] = *setupCmd.PersistentFlags().String("archive_command", "/bin/true", "The command to archive WAL files")
	pgSettings["archive_mode"] = *setupCmd.PersistentFlags().String("archive_mode", "on", "The archive mode (should be on to archive)")
	pgSettings["wal_level"] = *setupCmd.PersistentFlags().String("wal_level", "hot_standby", "The level of information to include in WAL files")
	//	setupCmd.PersistentFlags().String("", "", "")
	//	setupCmd.PersistentFlags().String("", "", "")
}

func check(err error) error {
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func getPgSetting(db *sql.DB, setting string) (value string, err error) {
	query := "SELECT setting FROM pg_settings WHERE name = $1;"
	rows, err := db.Query(query, setting)
	check(err)
	for rows.Next() {
		err = rows.Scan(&value)
		if err != nil {
			log.Fatal("Can't get PostgreSQL setting: ", setting, err)
			return "", err
		}
	}
	return value, nil
}

func setPgSetting(db *sql.DB, setting string, value string) (err error) {
	// Bad style and risk for injection!!! But not better option ...
	query := "ALTER SYSTEM SET " + setting + " = '" + value + "';"
	_, err = db.Query(query)
	if err != nil {
		log.Fatal("Can't set PostgreSQL setting: ", setting, " to: ", value, " Error: ", err)
		return err
	}
	log.Info("Set PostgreSQL setting: ", setting, " to: ", value)
	return nil
}

func configurePostgreSQL(db *sql.DB, settings map[string]string) (changed int, err error) {
	for setting := range settings {
		changed := 0
		settingShould := settings[setting]
		settingIs, err := getPgSetting(db, setting)
		check(err)
		log.Debug(setting, " should be: ", settingShould, " it is: ", settingIs)

		if settingIs != settingShould {
			err := setPgSetting(db, setting, settingShould)
			check(err)
			changed++
		}

	}
	return changed, nil
}

func testTools(tools []string) {
	for _, tool := range tools {
		cmd := exec.Command("command", "-v", tool)
		err := cmd.Run()
		check(err)
	}
}

func createDirs(archivedir string) error {
	dirs := []string{"current", "base", "wal"}
	for _, dir := range dirs {
		path := archivedir + "/" + dir
		err := os.MkdirAll(path, 0700)
		if err != nil {
			log.Fatal("Can not create directory: ", path)
			return err
		}
	}
	return nil
}
