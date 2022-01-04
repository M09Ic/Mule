package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// BruteCmd represents the Brute command
var CrawlCmd = &cobra.Command{
	Use:   "Crawl",
	Short: "crawl the link from the page",
	Long:  `I'm too lazy to write more introduction`,
	RunE:  StartCrawl,
}

func init() {
	rootCmd.AddCommand(CrawlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// BruteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// BruteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	CrawlCmd.Flags().StringP("url", "u", "", "brute target(currently only single url)")
	CrawlCmd.Flags().StringP("urls", "U", "", "targets from file")
	CrawlCmd.Flags().StringP("output", "o", "./res.log", "output res default in ./res.log")
	CrawlCmd.Flags().StringArrayP("Headers", "H", []string{}, "Request's Headers")
	CrawlCmd.Flags().StringP("Cookie", "C", "", "Request's Cookie")
	CrawlCmd.Flags().IntP("timeout", "", 5, "request's timeout")
	CrawlCmd.Flags().IntP("Thread", "t", 30, "the size of thread pool")
}

func StartCrawl(cmd *cobra.Command, args []string) error {
	fmt.Println("Hello Crawl")
	return nil
}
