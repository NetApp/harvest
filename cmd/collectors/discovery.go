package collectors

import (
	"errors"
	"slices"
	"time"

	ciscorest "github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	eseriesrest "github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	sgrest "github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
)

func GatherClusterInfo(pollerName string, cred *auth.Credentials, cols []conf.Collector) (conf.Remote, error) {
	// If the customer does not have a ZAPI collector, skip the checkZapi call
	hasZapi := slices.ContainsFunc(cols, func(c conf.Collector) bool {
		return c.Name == "Zapi" || c.Name == "ZapiPerf"
	})

	if !hasZapi {
		remote, err := checkRest(pollerName, cred)
		if err == nil {
			remote.ZAPIsChecked = false
		}
		return remote, err
	}

	remoteZapi, errZapi := checkZapi(pollerName, cred)
	remoteRest, errRest := checkRest(pollerName, cred)

	return MergeRemotes(remoteZapi, remoteRest, errZapi, errRest)
}

func GatherCiscoSwitchInfo(pollerName string, cred *auth.Credentials) (conf.Remote, error) {
	return checkCiscoRest(pollerName, cred)
}

func GatherStorageGridInfo(pollerName string, cred *auth.Credentials) (conf.Remote, error) {
	return checkStorageGrid(pollerName, cred)
}

func GatherEseriesInfo(pollerName string, cred *auth.Credentials) (conf.Remote, error) {
	return checkEseries(pollerName, cred)
}

func MergeRemotes(remoteZapi conf.Remote, remoteRest conf.Remote, errZapi error, errRest error) (conf.Remote, error) {
	remoteRest.ZAPIsExist = remoteZapi.ZAPIsExist
	remoteRest.ZAPIsChecked = remoteZapi.ZAPIsChecked

	// If both failed, return the combined error
	if errZapi != nil && errRest != nil {
		return remoteRest, errors.Join(errZapi, errRest)
	}

	// If at least one succeeded, return no error
	// Prefer REST remote if available, otherwise use ZAPI remote
	if errRest == nil {
		return remoteRest, nil
	}

	// errRest != nil but errZapi == nil (only ZAPI succeeded)
	return remoteZapi, nil
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

	if err := client.Init(1, conf.Remote{}); err != nil {
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
	err = client.Init(1, conf.Remote{})

	if err != nil {

		returnErr := true

		if he, ok := errors.AsType[errs.HarvestError](err); ok {
			switch {
			case he.ErrNum == errs.ErrNumZAPISuspended, he.StatusCode >= 400 && he.StatusCode < 500:
				// ZAPI is suspended, or we got a 4xx error, so we assume that ZAPI is not available
				zapisExist = false
				returnErr = false
			}
		}

		if returnErr {
			// Assume that ZAPIs exist so we don't upgrade ZAPI to REST when there is an error
			return conf.Remote{ZAPIsExist: true, ZAPIsChecked: true}, err
		}
	}

	remote := client.Remote()
	remote.ZAPIsChecked = true
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

	if err := client.Init(1, conf.Remote{}); err != nil {
		return conf.Remote{}, err
	}

	return client.Remote(), nil
}

func checkStorageGrid(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		poller *conf.Poller
		client *sgrest.Client
		err    error
	)

	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return conf.Remote{}, err
	}

	timeout, _ := time.ParseDuration(sgrest.DefaultTimeout)
	client, err = sgrest.New(poller, timeout, cred)
	if err != nil {
		return conf.Remote{}, err
	}

	if err := client.Init(1, conf.Remote{}); err != nil {
		return conf.Remote{}, err
	}

	return client.Remote, nil
}

func checkEseries(pollerName string, cred *auth.Credentials) (conf.Remote, error) {

	var (
		poller *conf.Poller
		client *eseriesrest.Client
		err    error
	)

	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return conf.Remote{}, err
	}

	timeout, _ := time.ParseDuration(eseriesrest.DefaultTimeout)
	client, err = eseriesrest.New(poller, timeout, cred, "")
	if err != nil {
		return conf.Remote{}, err
	}

	if err := client.Init(1, conf.Remote{}); err != nil {
		return conf.Remote{}, err
	}

	return client.Remote(), nil
}
