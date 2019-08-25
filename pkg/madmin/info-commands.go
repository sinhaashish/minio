/*
 * MinIO Cloud Storage, (C) 2017 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package madmin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio/pkg/cpu"
	"github.com/minio/minio/pkg/disk"
	"github.com/minio/minio/pkg/event/target"
	"github.com/minio/minio/pkg/mem"
)

// BackendType - represents different backend types.
type BackendType int

// Enum for different backend types.
const (
	Unknown BackendType = iota
	// Filesystem backend.
	FS
	// Multi disk Erasure (single, distributed) backend.
	Erasure

	// Add your own backend.
)

// DriveInfo - represents each drive info, describing
// status, uuid and endpoint.
type DriveInfo HealDriveInfo

// StorageInfo - represents total capacity of underlying storage.
type StorageInfo struct {
	Used uint64 // Total used spaced per tenant.

	Available uint64 // Total available space.

	Total uint64 // Total disk space.

	// Backend type.
	Backend struct {
		// Represents various backend types, currently on FS and Erasure.
		Type BackendType

		// Following fields are only meaningful if BackendType is Erasure.
		OnlineDisks      int // Online disks during server startup.
		OfflineDisks     int // Offline disks during server startup.
		StandardSCData   int // Data disks for currently configured Standard storage class.
		StandardSCParity int // Parity disks for currently configured Standard storage class.
		RRSCData         int // Data disks for currently configured Reduced Redundancy storage class.
		RRSCParity       int // Parity disks for currently configured Reduced Redundancy storage class.

		// List of all disk status, this is only meaningful if BackendType is Erasure.
		Sets [][]DriveInfo
	}
}

// ServerProperties holds some of the server's information such as uptime,
// version, region, ..
type ServerProperties struct {
	Uptime       time.Duration `json:"uptime"`
	Version      string        `json:"version"`
	CommitID     string        `json:"commitID"`
	DeploymentID string        `json:"deploymentID"`
	Region       string        `json:"region"`
	SQSARN       []string      `json:"sqsARN"`
}

// ServerConnStats holds network information
type ServerConnStats struct {
	TotalInputBytes  uint64 `json:"transferred"`
	TotalOutputBytes uint64 `json:"received"`
}

// ServerHTTPMethodStats holds total number of HTTP operations from/to the server,
// including the average duration the call was spent.
type ServerHTTPMethodStats struct {
	Count       uint64 `json:"count"`
	AvgDuration string `json:"avgDuration"`
}

// ServerHTTPStats holds all type of http operations performed to/from the server
// including their average execution time.
type ServerHTTPStats struct {
	TotalHEADStats     ServerHTTPMethodStats `json:"totalHEADs"`
	SuccessHEADStats   ServerHTTPMethodStats `json:"successHEADs"`
	TotalGETStats      ServerHTTPMethodStats `json:"totalGETs"`
	SuccessGETStats    ServerHTTPMethodStats `json:"successGETs"`
	TotalPUTStats      ServerHTTPMethodStats `json:"totalPUTs"`
	SuccessPUTStats    ServerHTTPMethodStats `json:"successPUTs"`
	TotalPOSTStats     ServerHTTPMethodStats `json:"totalPOSTs"`
	SuccessPOSTStats   ServerHTTPMethodStats `json:"successPOSTs"`
	TotalDELETEStats   ServerHTTPMethodStats `json:"totalDELETEs"`
	SuccessDELETEStats ServerHTTPMethodStats `json:"successDELETEs"`
}

// ServerInfoData holds storage, connections and other
// information of a given server
type ServerInfoData struct {
	StorageInfo StorageInfo      `json:"storage"`
	ConnStats   ServerConnStats  `json:"network"`
	HTTPStats   ServerHTTPStats  `json:"http"`
	Properties  ServerProperties `json:"server"`
}

// ServerInfo holds server information result of one node
type ServerInfo struct {
	Error string          `json:"error"`
	Addr  string          `json:"addr"`
	Data  *ServerInfoData `json:"data"`
}

// ServerInfo - Connect to a minio server and call Server Info Management API
// to fetch server's information represented by ServerInfo structure
func (adm *AdminClient) ServerInfo() ([]ServerInfo, error) {
	v := url.Values{}
	v.Set("type", string("server"))
	resp, err := adm.executeMethod("GET", requestData{relPath: "/v1/info", queryValues: v})
	defer closeResponse(resp)
	if err != nil {
		return nil, err
	}

	// Check response http status code
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}

	// Unmarshal the server's json response
	var serversInfo []ServerInfo

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &serversInfo)
	if err != nil {
		return nil, err
	}

	return serversInfo, nil
}

// ServerDrivesPerfInfo holds informantion about address and write speed of
// all drives in a single server node
type ServerDrivesPerfInfo struct {
	Addr  string             `json:"addr"`
	Error string             `json:"error,omitempty"`
	Perf  []disk.Performance `json:"perf"`
}

// ServerDrivesPerfInfo - Returns drive's read and write performance information
func (adm *AdminClient) ServerDrivesPerfInfo() ([]ServerDrivesPerfInfo, error) {
	v := url.Values{}
	v.Set("perfType", string("drive"))
	resp, err := adm.executeMethod("GET", requestData{
		relPath:     "/v1/performance",
		queryValues: v,
	})

	defer closeResponse(resp)
	if err != nil {
		return nil, err
	}

	// Check response http status code
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}

	// Unmarshal the server's json response
	var info []ServerDrivesPerfInfo

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ServerCPULoadInfo holds information about address and cpu load of
// a single server node
type ServerCPULoadInfo struct {
	Addr         string     `json:"addr"`
	Error        string     `json:"error,omitempty"`
	Load         []cpu.Load `json:"load"`
	HistoricLoad []cpu.Load `json:"historicLoad"`
}

// ServerCPULoadInfo - Returns cpu utilization information
func (adm *AdminClient) ServerCPULoadInfo() ([]ServerCPULoadInfo, error) {
	v := url.Values{}
	v.Set("perfType", string("cpu"))
	resp, err := adm.executeMethod("GET", requestData{
		relPath:     "/v1/performance",
		queryValues: v,
	})

	defer closeResponse(resp)
	if err != nil {
		return nil, err
	}

	// Check response http status code
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}

	// Unmarshal the server's json response
	var info []ServerCPULoadInfo

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ServerMemUsageInfo holds information about address and memory utilization of
// a single server node
type ServerMemUsageInfo struct {
	Addr          string      `json:"addr"`
	Error         string      `json:"error,omitempty"`
	Usage         []mem.Usage `json:"usage"`
	HistoricUsage []mem.Usage `json:"historicUsage"`
}

// ServerMemUsageInfo - Returns mem utilization information
func (adm *AdminClient) ServerMemUsageInfo() ([]ServerMemUsageInfo, error) {
	v := url.Values{}
	v.Set("perfType", string("mem"))
	resp, err := adm.executeMethod("GET", requestData{
		relPath:     "/v1/performance",
		queryValues: v,
	})

	defer closeResponse(resp)
	if err != nil {
		return nil, err
	}

	// Check response http status code
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}

	// Unmarshal the server's json response
	var info []ServerMemUsageInfo

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// Target has the list of targets with their status availibility info
type Target struct {
	Status        bool                      `json:"status,omitempty"`
	Error         error                     `json:"error,omitempty"`
	AMQP          *target.AMQPArgs          `json:"amqp,omitempty"`
	Elasticsearch *target.ElasticsearchArgs `json:"elasticsearch,omitempty"`
	Kafka         *target.KafkaArgs         `json:"kafka,omitempty"`
	MQTT          *target.MQTTArgs          `json:"mqtt,omitempty"`
	MySQL         *target.MySQLArgs         `json:"mysql,omitempty"`
	NATS          *target.NATSArgs          `json:"nats,omitempty"`
	NSQ           *target.NSQArgs           `json:"nsq,omitempty"`
	PostgreSQL    *target.PostgreSQLArgs    `json:"postgresql,omitempty"`
	Redis         *target.RedisArgs         `json:"redis,omitempty"`
	Web           *target.WebhookArgs       `json:"webhook,omitempty"`
}

// // AQMPInfo with status
// type AQMPInfo struct {
// 	Status bool            `json:"status"`
// 	Error  error           `json:"error"`
// 	AMQP   target.AMQPArgs `json:"amqp"`
// }

// // ElasticsearchInfo with status
// type ElasticsearchInfo struct {
// 	Status        bool                     `json:"status"`
// 	Error         error                    `json:"error"`
// 	Elasticsearch target.ElasticsearchArgs `json:"elasticsearch"`
// }

// // KafkaInfo with status
// type KafkaInfo struct {
// 	Status bool             `json:"status"`
// 	Error  error            `json:"error"`
// 	Kafka  target.KafkaArgs `json:"kafka"`
// }

// // MQTTInfo with status
// type MQTTInfo struct {
// 	Status bool            `json:"status"`
// 	Error  error           `json:"error"`
// 	MQTT   target.MQTTArgs `json:"mqtt"`
// }

// // MySQLInfo with status
// type MySQLInfo struct {
// 	Status bool             `json:"status"`
// 	Error  error            `json:"error"`
// 	MySQL  target.MySQLArgs `json:"mysql"`
// }

// // NATSInfo with status
// type NATSInfo struct {
// 	Status bool            `json:"status"`
// 	Error  error           `json:"error"`
// 	NATS   target.NATSArgs `json:"nats"`
// }

// // NSQInfo with status
// type NSQInfo struct {
// 	Status bool           `json:"status"`
// 	Error  error          `json:"error"`
// 	NSQ    target.NSQArgs `json:"nsq"`
// }

// // PostgreSQLInfo with status
// type PostgreSQLInfo struct {
// 	Status     bool                  `json:"status"`
// 	Error      error                 `json:"error"`
// 	PostgreSQL target.PostgreSQLArgs `json:"postgresql"`
// }

// // RedisInfo with status
// type RedisInfo struct {
// 	Status bool             `json:"status"`
// 	Error  error            `json:"error"`
// 	Redis  target.RedisArgs `json:"redis"`
// }

// // WebhookInfo with status
// type WebhookInfo struct {
// 	Status bool               `json:"status"`
// 	Error  error              `json:"error"`
// 	Web    target.WebhookArgs `json:"webhook"`
// }

// LambdaInfo -
type LambdaInfo struct {
	LambdaList []Target `json:"target"`
}

// ServerLambdaInfo fetches the logger server info
func (adm *AdminClient) ServerLambdaInfo() ([]Target, error) {
	v := url.Values{}
	v.Set("type", string("lambda"))
	resp, err := adm.executeMethod("GET", requestData{
		relPath:     "/v1/info",
		queryValues: v,
	})
	defer closeResponse(resp)
	if err != nil {
		return nil, err
	}

	// Check response http status code
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}

	// Unmarshal the server's json response
	// Unmarshal the server's json response
	var lambdaInfo []Target

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(respBytes, &lambdaInfo)
	if err != nil {
		return nil, err
	}

	return lambdaInfo, nil
}
