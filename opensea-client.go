package opensea

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	mainnetAPI = "https://api.opensea.io"
	rinkebyAPI = "https://rinkeby-api.opensea.io"
)

type Opensea struct {
	API        string
	APIKey     string
	httpClient *http.Client
}

type errorResponse struct {
	Success bool `json:"success" bson:"success"`
}

func (e errorResponse) Error() string {
	return "Not success"
}

func NewOpensea(apiKey string) (*Opensea, error) {
	o := &Opensea{
		API:        mainnetAPI,
		APIKey:     apiKey,
		httpClient: defaultHttpClient(),
	}
	return o, nil
}

func NewOpenseaRinkeby(apiKey string) (*Opensea, error) {
	o := &Opensea{
		API:        rinkebyAPI,
		APIKey:     apiKey,
		httpClient: defaultHttpClient(),
	}
	return o, nil
}

func (o Opensea) GetAssets(params GetAssetsParams) (*AssetsResponse, error) {
	ctx := context.TODO()
	return o.GetAssetsWithContext(ctx, params)
}

func (o Opensea) GetAssetsWithContext(ctx context.Context, params GetAssetsParams) (*AssetsResponse, error) {
	path := fmt.Sprintf("/api/v1/assets")
	values := url.Values{}
	if params.Owner != "" {
		values.Set("owner", params.Owner.String())
	}
	if len(params.TokenIds) > 0 {
		for _, tokenId := range params.TokenIds {
			values.Add("token_id", tokenId)
		}
	}
	if params.Collection != "" {
		values.Set("collection", params.Collection)
	}
	if params.CollectionSlug != "" {
		values.Set("collection_slug", params.CollectionSlug)
	}
	if params.CollectionEditor != "" {
		values.Set("collection_editor", params.CollectionEditor)
	}
	if params.OrderDirection != "" {
		values.Set("order_direction", string(params.OrderDirection))
	}
	if params.AssetContractAddress != "" {
		values.Set("asset_contract_address", params.AssetContractAddress.String())
	}
	if len(params.AssetContractAddresses) > 0 {
		for _, assetContractAddress := range params.AssetContractAddresses {
			values.Add("asset_contract_addresses", assetContractAddress.String())
		}
	}
	if params.Limit != 0 {
		values.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Cursor != "" {
		values.Set("cursor", params.Cursor)
	}
	if params.IncludeOrders {
		values.Set("include_orders", "true")
	}

	b, err := o.GetPath(ctx, path + values.Encode())
	if err != nil {
		return nil, err
	}
	ret := new(AssetsResponse)
	return ret, json.Unmarshal(b, ret)
}

func (o Opensea) GetSingleAsset(assetContractAddress string, tokenID *big.Int) (*Asset, error) {
	ctx := context.TODO()
	return o.GetSingleAssetWithContext(ctx, assetContractAddress, tokenID)
}

func (o Opensea) GetSingleAssetWithContext(ctx context.Context, assetContractAddress string, tokenID *big.Int) (
	*Asset,
	error,
) {
	path := fmt.Sprintf("/api/v1/asset/%s/%s", assetContractAddress, tokenID.String())
	b, err := o.GetPath(ctx, path)
	if err != nil {
		return nil, err
	}
	ret := new(Asset)
	return ret, json.Unmarshal(b, ret)
}

func (o Opensea) GetPath(ctx context.Context, path string) ([]byte, error) {
	return o.getURL(ctx, o.API+path)
}

func (o Opensea) getURL(ctx context.Context, url string) ([]byte, error) {
	client := o.httpClient
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Add("X-API-KEY", o.APIKey)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		e := new(errorResponse)
		err = json.Unmarshal(body, e)
		if err != nil {
			return nil, err
		}
		if !e.Success {
			return nil, e
		}

		return nil, fmt.Errorf("Backend returns status %d msg: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (o Opensea) SetHttpClient(httpClient *http.Client) {
	o.httpClient = httpClient
}

func defaultHttpClient() *http.Client {
	client := new(http.Client)
	var transport http.RoundTripper = &http.Transport{
		Proxy:              http.ProxyFromEnvironment,
		DisableKeepAlives:  false,
		DisableCompression: false,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 300 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client.Transport = transport
	return client
}
