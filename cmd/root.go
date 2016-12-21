// Copyright © 2016 Amanpreet Singh <aps.sids@gmail.com>
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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/spf13/cobra"
)

var queueName string
var url string
var awsRegion string
var sqsEndpoint string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gohaqd",
	Short: "gohaqd is a queue consuming worker daemon",
	Long: `A worker that pulls data off a queue, inserts it into the message body
and sends an HTTP POST request to a user-configurable URL.`,
	Run: startGohaqd,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&queueName, "queue-name", "q", "", "queue name to use")
	RootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "endpoint to send an HTTP POST request with contents of queue message in the body")
	RootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "us-east-1", "AWS Region for the SQS queue")
	RootCmd.PersistentFlags().StringVar(&sqsEndpoint, "sqs-endpoint", "", "SQS Endpoint for using with fake_sqs")
	RootCmd.MarkPersistentFlagRequired("queuename")
	RootCmd.MarkPersistentFlagRequired("url")
}

var svc *sqs.SQS
var msgparams *sqs.ReceiveMessageInput

func startGohaqd(cmd *cobra.Command, args []string) {
	var config *aws.Config
	if sqsEndpoint != "" {
		config = aws.NewConfig().WithEndpoint(sqsEndpoint).WithRegion(awsRegion)
	} else {
		config = aws.NewConfig().WithRegion(awsRegion)
	}
	sess := session.New(config)
	svc = sqs.New(sess)

	qparams := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	}

	q, err := svc.GetQueueUrl(qparams)
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("Polling SQS queue '%s' indefinitely..\n", queueName)
	msgparams = &sqs.ReceiveMessageInput{
		QueueUrl:        q.QueueUrl,
		WaitTimeSeconds: aws.Int64(20),
	}
	for {
		pollSQS(q.QueueUrl)
	}
}

func pollSQS(queueURL *string) {
	resp, err := svc.ReceiveMessage(msgparams)
	if err != nil {
		log.Fatalf(err.Error())
	}

	for _, msg := range resp.Messages {
		if sendMessageToURL(*msg.Body) {
			_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Fatalf(err.Error())
			}
		}
	}
}

func sendMessageToURL(msg string) bool {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(msg)))

	req.Header.Set("User-Agent", "gohaqd/0.2")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf(err.Error())
		return false
	}
	defer resp.Body.Close()

	// return true only if response is 200 OK
	if resp.StatusCode != 200 {
		log.Printf("Error: Non OK response. Status Code: '%s' for sent message: '%s'", resp.Status, msg)
		return false
	}

	return true
}
