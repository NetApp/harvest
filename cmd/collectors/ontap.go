package collectors

import (
	"errors"
	ciscorest "github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"net/http"
	"strings"
	"time"
)

func GatherClusterInfo(pollerName string, cred *auth.Credentials, cols []conf.Collector) (conf.Remote, error) {

	for _, col := range cols {
		if strings.HasPrefix(col.Name, "Cisco") {
			return checkCiscoRest(pollerName, cred)
		}
	}

	remoteZapi, errZapi := checkZapi(pollerName, cred)
	remoteRest, errRest := checkRest(pollerName, cred)

	return MergeRemotes(remoteZapi, remoteRest, errZapi, errRest)
}

func MergeRemotes(remoteZapi conf.Remote, remoteRest conf.Remote, errZapi error, errRest error) (conf.Remote, error) {
	err := errors.Join(errZapi, errRest)

	remoteRest.ZAPIsExist = remoteZapi.ZAPIsExist

	if errZapi != nil {
		return remoteRest, err
	}

	if errRest != nil {
		return remoteZapi, err
	}

	return remoteRest, err
}

func checkRest(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		poller *conf.Poller
		client *rest.Client
		err    error
	)

	// connect to the cluster
	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return conf.Remote{}, err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	client, err = rest.New(poller, timeout, cred)
	if err != nil {
		return conf.Remote{}, err
	}

	if err := client.Init(5, conf.Remote{}); err != nil {
		return conf.Remote{}, err
	}

	return client.Remote(), nil
}

func checkZapi(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		poller     *conf.Poller
		client     *zapi.Client
		err        error
		zapisExist bool
	)

	// connect to the cluster and retrieve the system version
	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return conf.Remote{}, err
	}
	if client, err = zapi.New(poller, cred); err != nil {
		return conf.Remote{}, err
	}

	zapisExist = true
	err = client.Init(2, conf.Remote{})

	if err != nil {

		returnErr := true

		var he errs.HarvestError
		if errors.As(err, &he) {
			switch {
			case he.ErrNum == errs.ErrNumZAPISuspended, he.StatusCode == http.StatusBadRequest:
				zapisExist = false
				returnErr = false
			}
		}

		if returnErr {
			// Assume that ZAPIs exist so we don't upgrade ZAPI to REST when there is an error
			return conf.Remote{ZAPIsExist: true}, err
		}
	}

	remote := client.Remote()
	remote.ZAPIsExist = zapisExist

	return remote, nil
}

func checkCiscoRest(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		poller *conf.Poller
		client *ciscorest.Client
		err    error
	)

	// connect to the switch
	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return conf.Remote{}, err
	}

	timeout, _ := time.ParseDuration(ciscorest.DefaultTimeout)
	client, err = ciscorest.New(poller, timeout, cred)
	if err != nil {
		return conf.Remote{}, err
	}

	if err := client.Init(5, conf.Remote{}); err != nil {
		return conf.Remote{}, err
	}

	return client.Remote(), nil
}
