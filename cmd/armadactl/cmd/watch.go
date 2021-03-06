package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/G-Research/armada/internal/armada/api"
	"github.com/G-Research/armada/internal/client"
	"github.com/G-Research/armada/internal/client/domain"
	"github.com/G-Research/armada/internal/client/util"
)

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().Bool("raw", false, "Output raw events")
}

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch job events in job set.",
	Long:  `This command will list all job set events and `,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		jobSetId := args[0]
		raw, _ := cmd.Flags().GetBool("raw")

		log.Infof("Watching job set %s", jobSetId)

		apiConnectionDetails := client.ExtractCommandlineArmadaApiConnectionDetails()

		util.WithConnection(apiConnectionDetails, func(conn *grpc.ClientConn) {
			eventsClient := api.NewEventClient(conn)
			client.WatchJobSet(eventsClient, jobSetId, true, context.Background(), func(state *domain.WatchContext, e api.Event) bool {
				if raw {
					data, err := json.Marshal(e)
					if err != nil {
						log.Error(e)
					} else {
						log.Infof("%s %s\n", reflect.TypeOf(e), string(data))
					}
				} else {
					summary := fmt.Sprintf("%s | ", e.GetCreated().Format(time.Stamp))
					summary += state.GetCurrentStateSummary()
					summary += fmt.Sprintf(" | event: %s, job id: %s", reflect.TypeOf(e), e.GetJobId())
					log.Info(summary)
				}
				return false
			})
		})
	},
}
