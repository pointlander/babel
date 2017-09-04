// Copyright 2017 The Babel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"compress/bzip2"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/davecheney/profile"
)

var skip_list = []string{
	//problems with bold and italic - fixed
	"Aragonese language",         //slow time
	"Catalan language",           //slow time
	"Docklands Light Railway",    //slow time
	"Family name",                //slow time
	"Fulham F.C.",                //slow time
	"List of FIPS country codes", //slow time
	"Grammatical case",           //slow time
	"Grammatical conjugation",    //slow time
	"Grazia Deledda",             //slow time
	"Höðr",                       //slow time
	"List of marine aquarium fish species", //slow time
	"Peptide",                                       //slow time
	"Patrick Macnee",                                //slow time
	"Quechuan languages",                            //slow time
	"List of Russian-language poets",                //slow time
	"Spanish language",                              //slow time
	"Semitic languages",                             //slow time
	"Septuagint",                                    //slow time
	"Sonnet",                                        //slow time
	"Steve Jackson Games",                           //slow time
	"Split infinitive",                              //slow time
	"The Goon Show",                                 //slow time
	"Work breakdown structure",                      //slow time
	"Pericles",                                      //slow time
	"Thomas Telford",                                //slow time
	"Planner (programming language)",                //slow time
	"Article (grammar)",                             //slow time
	"Hedd Wyn",                                      //slow time
	"Pierrot",                                       //slow time
	"Presidential Succession Act",                   //slow time
	"Women's National Basketball Association",       //slow time
	"List of popes",                                 //slow time
	"Academy Award for Best Picture",                //slow time
	"Academy Award for Best Live Action Short Film", //slow time

	//problems with table - fixed
	"Hyperinflation", //slow time

	//problems with invalid templates and section parsing
	"Wikipedia:Upload log archive/March 2003",                                 //slow time
	"Wikipedia:Upload log archive/April 2003",                                 //slow time
	"Wikipedia:Upload log archive/May 2003",                                   //slow time
	"Wikipedia:Upload log archive/June 2003",                                  //slow time
	"Wikipedia:People by year/Reports/No other categories/2",                  //slow time
	"Wikipedia:People by year/Reports/Canadians/For years in Canada (births)", //slow time
	"Wikipedia:People by year/Reports/Canadians/All",                          //slow time
	"Wikipedia:WikiProject Missing encyclopedic articles/Misc",                //slow time
	"Wikipedia:WikiProject Texas/Articles/Page3",
	"Wikipedia:WikiProject Automobiles/Articles/Page3",
	"Wikipedia:WikiProject Louisville/Watchall",
	"Wikipedia:WikiProject Kentucky/Watchall",
	"Wikipedia:Deletion log archive/August 2003",
	"Wikipedia:Deletion log archive/September 2003",
	"Wikipedia:WikiProject Iowa/Iowa recent changes",
	"Wikipedia:WikiProject Baseball/Articles/Page3",
	"Wikipedia:WikiProject Spam/LinkSearch/amazon.com",
	"Wikipedia:WikiProject Spam/COIReports/2007, Apr 26",
	"Wikipedia:WikiProject Spam/COIReports/2007, May 3",
	"Wikipedia:WikiProject Spam/COIReports/2007, May 16",
	"Wikipedia:WikiProject Spam/COIReports/2007, May 22",
	"Wikipedia:WikiProject Film/Articles/Page6",
	"Wikipedia:WikiProject Film/Articles/Page7",
	"Wikipedia:WikiProject Spam/COIReports/2007, May 23",
	"Wikipedia:WikiProject Spam/COIReports/2007, Jun 5",
	"Wikipedia:WikiProject Spam/COIReports/2007, Jun 29",
	"Wikipedia:WikiProject Spam/LinkReports/cia.gov",
	"Wikipedia:WikiProject Spam/LinkReports/web.archive.org",
	"Wikipedia:WikiProject Spam/LinkReports/pbs.org",
	"Wikipedia:WikiProject Spam/LinkReports/cnn.com",
	"Wikipedia:WikiProject Spam/LinkReports/un.org",
	"Wikipedia:WikiProject National Register of Historic Places/coordsA",

	//new problems
	"Template:NBA team standings",
	"Wikipedia:WikiProject Chemicals/Log/2008-11-17",
	"Wikipedia:WikiProject Chemicals/Log/2008-11-25",
	"Wikipedia:WikiProject Spam/LinkReports/mybrute.com",
	"Wikipedia:WikiProject Chemicals/Log/2009-04-30",
	"Wikipedia:WikiProject Chemicals/Log/2009-07-31",
}

var wiki_file = flag.String("file", "", "file to parse")
var parse_skips = flag.Bool("skips", false, "parse all skip files")

func main() {
	pro := profile.Start(profile.CPUProfile)
	defer pro.Stop()

	skip_file := func(currentTitle string) string {
		return "skip/" + strings.Replace(strings.Replace(currentTitle, " ", "_", -1), "/", "_", -1) + ".txt"
	}

	flag.Parse()
	if *wiki_file != "" {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("\t", r)
			}
		}()

		wiki_text, err := ioutil.ReadFile(*wiki_file)
		if err != nil {
			log.Fatal(err)
		}
		parser := &Wiki{Buffer: string(wiki_text), start_of_line: true}
		parser.Init()
		if err := parser.Parse(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *parse_skips {
		parse := func(name string) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("\t", r)
				}
			}()

			file_name := skip_file(name)
			fmt.Println(file_name)
			wiki_text, err := ioutil.ReadFile(file_name)
			if err != nil {
				log.Fatal(err)
			}
			parser := &Wiki{Buffer: string(wiki_text), start_of_line: true}
			parser.Init()
			if err := parser.Parse(); err != nil {
				log.Fatal(err)
			}
		}

		for _, j := range skip_list {
			parse(j)
		}
	}

	skip := make(map[string]interface{})
	for _, j := range skip_list {
		skip[j] = struct{}{}
	}

	out, err := os.Create("skip_list.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	file, err := os.Open("enwiki-latest-pages-articles.xml.bz2")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	parse := func(article, currentTitle string) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("\t", r)
			}
		}()

		//fmt.Printf("%v\n", article)
		start, slow_time, bad_parse := time.Now(), false, false
		parser := &Wiki{Buffer: article, start_of_line: true}
		parser.Init()
		if err := parser.Parse(); err != nil {
			bad_parse = true
		}
		if time.Now().Sub(start) > time.Minute {
			slow_time = true
		}
		if bad_parse || slow_time {
			out.WriteString(currentTitle)
			if bad_parse {
				out.WriteString(" bad parse")
			}
			if slow_time {
				out.WriteString(" slow time")
			}
			out.WriteString("\n")
		}
	}

	decoder := xml.NewDecoder(bzip2.NewReader(file))
	decoder.Strict = false
	inText, inTitle, title, article, currentTitle := false, false, "", "", ""
	for token, err := decoder.RawToken(); err == nil; token, err = decoder.RawToken() {
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "text" {
				inText = true
			} else if t.Name.Local == "title" {
				inTitle = true
			}
		case xml.CharData:
			if inText {
				article += string(t)
			} else if inTitle {
				title += string(t)
			}
		case xml.EndElement:
			if inText {
				if _, s := skip[currentTitle]; !s {
					parse(article, currentTitle)
				} else {
					skip_out, err := os.Create(skip_file(currentTitle))
					if err != nil {
						log.Fatal(err)
					}
					skip_out.WriteString(article)
					skip_out.Close()
				}
				inText, article = false, ""
			} else if inTitle {
				fmt.Printf("%v\n", title)
				currentTitle = title
				inTitle, title = false, ""
			}
		}
	}
}
