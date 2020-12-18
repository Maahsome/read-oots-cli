/*
Copyright © 2020 Christopher Maahs <cmaahs@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const baseURL = "https://www.giantitp.com"
const catalogURL = "https://www.giantitp.com/comics/oots.html"
const patreonURL = "https://www.patreon.com/oots"

var cfgFile string
var workDir string

type strip struct {
	Title  string
	Number string
	Link   string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "read-oots-cli",
	Short: "Read OOTS Comic Strips in the Terminal Window",
	Long: `In order to keep with with the pace at which new OOTS
	Comics come out, reading them in the terminal window is a big help.`,
	Run: func(cmd *cobra.Command, args []string) {
		stripNumber, _ := cmd.Flags().GetString("number")
		goNext, _ := cmd.Flags().GetBool("next")
		goPrevious, _ := cmd.Flags().GetBool("previous")
		setCurrent, _ := cmd.Flags().GetBool("set-current")

		// We default to goNext, these are times we don't want that
		if len(stripNumber) > 0 || goPrevious {
			goNext = false
		}

		num := getStripNumber(stripNumber, goNext, goPrevious)

		strips := collect(catalogURL)

		if val, ok := strips[num]; ok {
			wg := sync.WaitGroup{}
			wg.Add(1)
			download(val.Title, val.Number, val.Link, &wg)
			wg.Wait()
			mdcat := exec.Command("mdcat", fmt.Sprintf("%s/strip.md", workDir))
			mdcat.Stdout = os.Stdout
			mdcat.Run()
			if setCurrent || goPrevious || goNext {
				saveCurrentStrip(num)
			}
		}
	},
}

func getStripNumber(num string, next bool, prev bool) string {

	if len(num) > 0 {
		return num
	}

	num = getLastReadStrip()
	if len(num) == 0 {
		return "1"
	}
	i, converr := strconv.Atoi(num)
	if converr != nil {
		logrus.Warn("Why couldn't I convert this to a number? ", num)
	}

	if next {
		i++
	}
	if prev {
		i--
	}
	return fmt.Sprintf("%d", i)
}

func getLastReadStrip() string {
	return viper.GetString("current")
}

func saveCurrentStrip(num string) {
	viper.Set("current", num)
	verr := viper.WriteConfig()
	if verr != nil {
		logrus.WithError(verr).Info("Failed to write config")
	}
}

// Execute - Run the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.read-oots-cli.yaml)")

	rootCmd.Flags().StringP("number", "s", "", "Show a (S)pecific numbered comic strip")
	rootCmd.Flags().BoolP("next", "n", true, "Show the NEXT comic strip")
	rootCmd.Flags().BoolP("previous", "p", false, "Show the PREVIOUS comic strip")
	rootCmd.Flags().BoolP("set-current", "c", false, "Set the viewed comic # as the last viewed")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		workDir = fmt.Sprintf("%s/.config/read-oots-cli", home)
		if _, err := os.Stat(workDir); err != nil {
			if os.IsNotExist(err) {
				mkerr := os.MkdirAll(workDir, os.ModePerm)
				if mkerr != nil {
					logrus.Fatal("Error creating ~/.config/read-oots-cli directory", mkerr)
				}
			}
		}
		if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
			configFile := fmt.Sprintf("%s/%s", workDir, "config.yml")
			createRestrictedConfigFile(configFile)
			viper.SetConfigFile(configFile)
		} else {
			logrus.Info("The ~/.config/read-oots-cli path is a file and not a directory, please remove the 'read-oots-cli' file.")
			os.Exit(1)
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// logrus.Warn("Failed to read config file: ", viper.ConfigFileUsed())
	}
}

func createRestrictedConfigFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			file, ferr := os.Create(fileName)
			if ferr != nil {
				logrus.Info("Unable to create the configfile.")
				os.Exit(1)
			}
			mode := int(0600)
			if cherr := file.Chmod(os.FileMode(mode)); cherr != nil {
				logrus.Info("Chmod for config file failed, please set the mode to 0600.")
			}
		}
	}
}

func collect(url string) map[string]strip {
	c := colly.NewCollector(
		colly.AllowedDomains("www.giantitp.com"),
	)

	var strips = make(map[string]strip)

	c.OnHTML("p[class=ComicList]", func(e *colly.HTMLElement) {
		titleNumber := e.Text
		re, _ := regexp.Compile(`(\d+) - (.*)$`)
		result := re.FindAllStringSubmatch(titleNumber, -1)
		title := result[0][2]
		number := result[0][1]
		relLink := e.ChildAttr("a", "href")
		link := baseURL + relLink
		// strips = append(strips, strip{Title: title, Number: number, Link: link})
		strips[number] = strip{Title: title, Number: number, Link: link}
	})

	c.Visit(url)

	return strips
}

func download(title, number, url string, wg *sync.WaitGroup) {
	c := colly.NewCollector()

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		matched, _ := regexp.MatchString(`comics\/oots`, link)
		if matched == true {
			re := regexp.MustCompile(`com//`)
			link = re.ReplaceAllString(link, "com/")
			go func() {
				defer wg.Done()
				originalTitle := title
				extension := path.Ext(link)
				filename := workDir + "/strip" + extension
				response, err := http.Get(link)
				if err != nil {
					log.Fatal(err)
				}
				defer response.Body.Close()

				file, err := os.Create(filename)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()
				_, err = io.Copy(file, response.Body)
				if err != nil {
					log.Fatal(err)
				}

				mdFilename := workDir + "/strip" + ".md"
				mdown := fmt.Sprintf("# oots %s\n\n[%s](%s)\nSupport OotS on [Patreon](%s)\n\n![%s](%s)", number, originalTitle, url, patreonURL, number, filename)
				mdReader := strings.NewReader(mdown)
				mdfile, mderr := os.Create(mdFilename)
				if mderr != nil {
					log.Fatal(mderr)
				}
				defer mdfile.Close()
				_, err = io.Copy(mdfile, mdReader)
				if err != nil {
					log.Fatal(err)
				}

			}()
		}
	})

	c.Visit(url)

}
