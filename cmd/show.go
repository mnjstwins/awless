package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/badwolf/triple/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wallix/awless/cloud/aws"
	"github.com/wallix/awless/config"
	"github.com/wallix/awless/display"
	"github.com/wallix/awless/rdf"
	"github.com/wallix/awless/revision"
)

var (
	numberRevisionsToShow    int
	showRevisionsProperties  bool
	showRevisionsGroupAll    bool
	showRevisionsGroupByDay  bool
	showRevisionsGroupByWeek bool
)

func init() {
	//Resources
	for resource, properties := range display.PropertiesDisplayer.Services[aws.InfraServiceName].Resources {
		showCmd.AddCommand(showInfraResourceCmd(resource, properties))
	}
	for resource, properties := range display.PropertiesDisplayer.Services[aws.AccessServiceName].Resources {
		showCmd.AddCommand(showAccessResourceCmd(resource, properties))
	}

	//Revisions
	showCmd.AddCommand(showCloudRevisionsCmd)
	showCloudRevisionsCmd.PersistentFlags().IntVarP(&numberRevisionsToShow, "number", "n", 10, "Number of revision to show")
	showCloudRevisionsCmd.PersistentFlags().BoolVarP(&showRevisionsProperties, "properties", "p", false, "Full diff with resources properties")
	showCloudRevisionsCmd.PersistentFlags().BoolVar(&showRevisionsGroupAll, "group-all", false, "Group all revisions")
	showCloudRevisionsCmd.PersistentFlags().BoolVar(&showRevisionsGroupByWeek, "group-by-week", false, "Group revisions by week")
	showCloudRevisionsCmd.PersistentFlags().BoolVar(&showRevisionsGroupByDay, "group-by-day", false, "Group revisions by day")

	RootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show various type of items by id: users, groups, instances, vpcs, ...",
}

var showInfraResourceCmd = func(resource rdf.ResourceType, displayer *display.ResourceDisplayer) *cobra.Command {
	resources := pluralize(resource.String())
	command := &cobra.Command{
		Use:   resource.String() + " id",
		Short: "Show the properties of a AWS EC2 " + resource.String(),

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("id required")
			}
			id := args[0]
			var g *rdf.Graph
			var err error
			if localResources {
				g, err = rdf.NewGraphFromFile(filepath.Join(config.GitDir, config.InfraFilename))

			} else {
				g, err = fetchRemoteResource(aws.InfraService, resources)
			}
			exitOn(err)
			err = display.OneResourceOfGraph(os.Stdout, g, resource, id, displayer)
			exitOn(err)
			return nil
		},
	}

	command.PersistentFlags().BoolVar(&localResources, "local", false, "List locally sync resources")
	return command
}

var showAccessResourceCmd = func(resource rdf.ResourceType, displayer *display.ResourceDisplayer) *cobra.Command {
	resources := pluralize(resource.String())
	command := &cobra.Command{
		Use:   resource.String() + " id",
		Short: "Show the properties of a AWS IAM " + resource.String(),

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("id required")
			}
			id := args[0]
			var g *rdf.Graph
			var err error
			if localResources {
				g, err = rdf.NewGraphFromFile(filepath.Join(config.GitDir, config.AccessFilename))

			} else {
				g, err = fetchRemoteResource(aws.AccessService, resources)
			}
			exitOn(err)
			err = display.OneResourceOfGraph(os.Stdout, g, resource, id, displayer)
			exitOn(err)
			return nil
		},
	}

	command.PersistentFlags().BoolVar(&localResources, "local", false, "List locally sync resources")
	return command
}

var showCloudRevisionsCmd = &cobra.Command{
	Use:   "revisions",
	Short: "Show cloud revision history",

	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := node.NewNodeFromStrings("/region", viper.GetString("region"))
		if err != nil {
			return err
		}
		r, err := revision.OpenRepository(config.GitDir)
		if err != nil {
			return err
		}
		param := revision.NoGroup
		if showRevisionsGroupAll {
			param = revision.GroupAll
		}
		if showRevisionsGroupByDay {
			param = revision.GroupByDay
		}
		if showRevisionsGroupByWeek {
			param = revision.GroupByWeek
		}
		accessDiffs, err := r.LastDiffs(numberRevisionsToShow, root, param, config.AccessFilename)
		if err != nil {
			return err
		}
		infraDiffs, err := r.LastDiffs(numberRevisionsToShow, root, param, config.InfraFilename)
		if err != nil {
			return err
		}
		for i := range accessDiffs {
			display.RevisionDiff(accessDiffs[i], aws.AccessServiceName, root, verboseFlag, showRevisionsProperties)
			display.RevisionDiff(infraDiffs[i], aws.InfraServiceName, root, verboseFlag, showRevisionsProperties)
		}
		return nil
	},
}
