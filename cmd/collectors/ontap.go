package collectors

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"net/http"
	"time"
)

func GatherClusterInfo(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		err error
	)

	remoteZapi, err := checkZapi(pollerName, cred)
	if err != nil {
		return conf.Remote{}, err
	}

	remoteRest, err := checkRest(pollerName, cred)
	if err != nil {
		return remoteZapi, err
	}

	remoteRest.ZAPIsExist = remoteZapi.ZAPIsExist

	return remoteRest, nil
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
			return conf.Remote{}, err
		}
	}

	remote := client.Remote()
	remote.ZAPIsExist = zapisExist

	return remote, nil
}
