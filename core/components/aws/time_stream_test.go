package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestTimeStreamCRUD(t *testing.T) {
	//TimeStreamCRUD()

	// Setting 20 seconds for timeout
	tr := &http.Transport{
		ResponseHeaderTimeout: 20 * time.Second,
		// Using DefaultTransport values for other parameters: https://golang.org/pkg/net/http/#RoundTripper
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Timeout:   30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// So client makes HTTP/2 requests
	http2.ConfigureTransport(tr)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-northeast-1"), MaxRetries: aws.Int(10), HTTPClient: &http.Client{Transport: tr}})
	writeSvc := timestreamwrite.New(sess)

	databaseName := "MetricsDB"
	tableName := "Marketing"

	// Below code will ingest cpu_utilization and memory_utilization metric for a host on
	// region=us-east-1, az=az1, and hostname=host1

	// Get current time in seconds.
	now := time.Now()
	currentTimeInSeconds := now.Unix()
	err = write(databaseName, tableName, currentTimeInSeconds, err, writeSvc)

	version, writeRecordsCommonAttributesUpsertInput, err := firstWrite(now, currentTimeInSeconds, databaseName, tableName, err, writeSvc)

	err = retrySameWrites(err, writeSvc, writeRecordsCommonAttributesUpsertInput)

	update(writeRecordsCommonAttributesUpsertInput, version, err, writeSvc)

}

func write(databaseName string, tableName string, currentTimeInSeconds int64, err error, writeSvc *timestreamwrite.TimestreamWrite) error {
	writeRecordsInput := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
		Records: []*timestreamwrite.Record{
			&timestreamwrite.Record{
				Dimensions: []*timestreamwrite.Dimension{
					&timestreamwrite.Dimension{
						Name:  aws.String("region"),
						Value: aws.String("us-east-1"),
					},
					&timestreamwrite.Dimension{
						Name:  aws.String("az"),
						Value: aws.String("az1"),
					},
					&timestreamwrite.Dimension{
						Name:  aws.String("hostname"),
						Value: aws.String("host1"),
					},
				},
				MeasureName:      aws.String("cpu_utilization"),
				MeasureValue:     aws.String("13.5"),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             aws.String(strconv.FormatInt(currentTimeInSeconds, 10)),
				TimeUnit:         aws.String("SECONDS"),
			},
			&timestreamwrite.Record{
				Dimensions: []*timestreamwrite.Dimension{
					&timestreamwrite.Dimension{
						Name:  aws.String("region"),
						Value: aws.String("us-east-1"),
					},
					&timestreamwrite.Dimension{
						Name:  aws.String("az"),
						Value: aws.String("az1"),
					},
					&timestreamwrite.Dimension{
						Name:  aws.String("hostname"),
						Value: aws.String("host1"),
					},
				},
				MeasureName:      aws.String("memory_utilization"),
				MeasureValue:     aws.String("40"),
				MeasureValueType: aws.String("DOUBLE"),
				Time:             aws.String(strconv.FormatInt(currentTimeInSeconds, 10)),
				TimeUnit:         aws.String("SECONDS"),
			},
		},
	}

	_, err = writeSvc.WriteRecords(writeRecordsInput)

	if err != nil {
		fmt.Println("Error:")
		fmt.Println(err)
	} else {
		fmt.Println("Write records is successful")
	}
	return err
}

func firstWrite(now time.Time, currentTimeInSeconds int64, databaseName string, tableName string, err error, writeSvc *timestreamwrite.TimestreamWrite) (int64, *timestreamwrite.WriteRecordsInput, error) {
	// Below code will ingest and upsert cpu_utilization and memory_utilization metric for a host on
	// region=us-east-1, az=az1, and hostname=host1
	fmt.Println("Ingesting records and set version as currentTimeInMills, hit enter to continue")

	// Get current time in seconds.
	now = time.Now()
	currentTimeInSeconds = now.Unix()
	// To achieve upsert (last writer wins) semantic, one example is to use current time as the version if you are writing directly from the data source
	version := time.Now().Round(time.Millisecond).UnixNano() / 1e6 // set version as currentTimeInMills

	writeRecordsCommonAttributesUpsertInput := &timestreamwrite.WriteRecordsInput{
		DatabaseName: aws.String(databaseName),
		TableName:    aws.String(tableName),
		CommonAttributes: &timestreamwrite.Record{
			Dimensions: []*timestreamwrite.Dimension{
				&timestreamwrite.Dimension{
					Name:  aws.String("region"),
					Value: aws.String("us-east-1"),
				},
				&timestreamwrite.Dimension{
					Name:  aws.String("az"),
					Value: aws.String("az1"),
				},
				&timestreamwrite.Dimension{
					Name:  aws.String("hostname"),
					Value: aws.String("host1"),
				},
			},
			MeasureValueType: aws.String("DOUBLE"),
			Time:             aws.String(strconv.FormatInt(currentTimeInSeconds, 10)),
			TimeUnit:         aws.String("SECONDS"),
			Version:          &version,
		},
		Records: []*timestreamwrite.Record{
			&timestreamwrite.Record{
				MeasureName:  aws.String("cpu_utilization"),
				MeasureValue: aws.String("13.5"),
			},
			&timestreamwrite.Record{
				MeasureName:  aws.String("memory_utilization"),
				MeasureValue: aws.String("40"),
			},
		},
	}

	// write records for first time
	_, err = writeSvc.WriteRecords(writeRecordsCommonAttributesUpsertInput)

	if err != nil {
		fmt.Println("Error:")
		fmt.Println(err)
	} else {
		fmt.Println("Frist-time write records is successful")
	}
	return version, writeRecordsCommonAttributesUpsertInput, err
}

func retrySameWrites(err error, writeSvc *timestreamwrite.TimestreamWrite, writeRecordsCommonAttributesUpsertInput *timestreamwrite.WriteRecordsInput) error {
	fmt.Println("Retry same writeRecordsRequest with same records and versions. Because writeRecords API is idempotent, this will success. hit enter to continue")
	_, err = writeSvc.WriteRecords(writeRecordsCommonAttributesUpsertInput)
	return err
}

func update(writeRecordsCommonAttributesUpsertInput *timestreamwrite.WriteRecordsInput, version int64, err error, writeSvc *timestreamwrite.TimestreamWrite) {
	updated_cpu_utilization := &timestreamwrite.Record{
		MeasureName:  aws.String("cpu_utilization"),
		MeasureValue: aws.String("14.5"),
	}
	updated_memory_utilization := &timestreamwrite.Record{
		MeasureName:  aws.String("memory_utilization"),
		MeasureValue: aws.String("50"),
	}

	writeRecordsCommonAttributesUpsertInput.Records = []*timestreamwrite.Record{
		updated_cpu_utilization,
		updated_memory_utilization,
	}
	fmt.Println("Upsert with higher version as new data is generated, this would success. hit enter to continue")
	version = time.Now().Round(time.Millisecond).UnixNano() / 1e6 // set version as currentTimeInMills
	writeRecordsCommonAttributesUpsertInput.CommonAttributes.Version = &version
	writeRecordsCommonAttributesUpsertInput.Records = []*timestreamwrite.Record{
		&timestreamwrite.Record{
			MeasureName:  aws.String("cpu_utilization"),
			MeasureValue: aws.String("34.5"),
		},
		updated_memory_utilization,
	}

	_, err = writeSvc.WriteRecords(writeRecordsCommonAttributesUpsertInput)

	if err != nil {
		fmt.Println("Error:")
		fmt.Println(err)
	} else {
		fmt.Println("Write records with higher version is successful")
	}
}
