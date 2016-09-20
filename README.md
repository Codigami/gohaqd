# gohaqd

`gohaqd` _(pronounced as go-hawk-d)_ is a worker daemon for consuming messages from a queue.

It pulls data off a queue, inserts it into the message body, and sends an HTTP POST request to a user-configurable URL. Currently supports only AWS SQS.

## Usage:
```
  gohaqd [flags]
```
## Flags:
```
      --aws-region string   AWS Region for the SQS queue (default "us-east-1")
  -q, --queue-name string   queue name to use
  -u, --url string          endpoint to send an HTTP POST request with contents of queue message in the body
```
