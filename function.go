package email_func

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/itmayziii/email/send"
	"github.com/mailgun/mailgun-go/v4"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	"log"
	"os"
)

// init exists purely as an entry point for GCP Cloud functions.
// https://cloud.google.com/functions/docs/writing#entry-point
func init() {
	loggingClient, err := logging.NewClient(context.Background(), os.Getenv("PROJECT_ID"))
	if err != nil {
		log.Fatal("failed to create logging client", err)
	}
	logger := loggingClient.Logger("email", logging.RedirectAsJSON(os.Stdout))
	infoLogger := logger.StandardLogger(logging.Info)
	errorLogger := logger.StandardLogger(logging.Error)

	ctx := context.Background()
	bucketName := os.Getenv("BUCKET")
	// Do not worry about closing the bucket, this is a no-op for all the implementations that we use i.e. file, GCS, S3.
	// It would be more difficult to close the bucket because function invocations sharing the same cloud function would
	// not be able to share resources.
	bucket, err := blob.OpenBucket(ctx, bucketName)
	if err != nil {
		logger.LogSync(ctx, logging.Entry{
			Severity: logging.Alert,
			Payload:  fmt.Sprintf("failed to open bucket %s - %v", bucketName, err),
		})
		os.Exit(1)
	}

	app := send.NewApp(
		send.AppWithFlusher(logger),
		send.AppWithInfoLogger(infoLogger),
		send.AppWithErrorLogger(errorLogger),
		send.AppWithFileStorage(bucket),
		send.AppWithDomainSender("tommymay.dev", send.NewMailgunSender(
			mailgun.NewMailgun("mg.tommymay.dev", os.Getenv("MG_API_KEY_MG_TOMMYMAY_DEV")),
		)),
	)
	functions.CloudEvent("Email", send.EmailEvent(app))
}
