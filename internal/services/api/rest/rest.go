package rest

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/crypto/ripemd160"
)

const (
	valPass = "/cosmos/staking/v1beta1/validators/"
)

type ValidatorResponse struct {
	Validator struct {
		OperatorAddress   string `json:"operator_address"`
		ConsensusPubkey   ConsensusPubkey `json:"consensus_pubkey"`
		Jailed            bool   `json:"jailed"`
		Status            string `json:"status"`
		Tokens            string `json:"tokens"`
		Description       Description `json:"description"`
		Commission        Commission `json:"commission"`
	} `json:"validator"`
}

type ConsensusPubkey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type Description struct {
	Moniker         string `json:"moniker"`
	Website         string `json:"website"`
	Details         string `json:"details"`
}

type Commission struct {
	CommissionRates struct {
		Rate          string `json:"rate"`
		MaxRate       string `json:"max_rate"`
		MaxChangeRate string `json:"max_change_rate"`
	} `json:"commission_rates"`
	UpdateTime string `json:"update_time"`
}

// Получение консенсус кея с валадреса
func getAddressConsensusKey(api, address string) (ConsensusPubkey, error) {
	apiurl, _ := url.JoinPath(api , valPass, address)
	response, err := http.Get(apiurl)
	if err != nil {
		fmt.Println("Ошибка при выполнении GET-запроса:", err)
		return ConsensusPubkey{}, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return ConsensusPubkey{}, err
	}

	var resp ValidatorResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("Ошибка при распаковке JSON:", err)
		return ConsensusPubkey{}, err
	}
	return resp.Validator.ConsensusPubkey, nil
}

func GetHexAddress(api, address string) (HexAddress string) {
	consensusPubkey, _ := getAddressConsensusKey(api, address)
	switch consensusPubkey.Type {
	case "/cosmos.crypto.ed25519.PubKey":
		pubkey, err := base64.StdEncoding.DecodeString(consensusPubkey.Key)
		if err != nil {
			fmt.Println("Error decoding base64:", err)
			return
		}

		hash := sha256.Sum256(pubkey)
		hexHash := hex.EncodeToString(hash[:])
		result := strings.ToUpper(hexHash[:40])

		return result
	case "/cosmos.crypto.secp256k1.PubKey":
		pubkey, err := base64.StdEncoding.DecodeString(consensusPubkey.Key)
		if err != nil {
			fmt.Println("Error decoding base64:", err)
			return
		}

		sha256Hash := sha256.Sum256(pubkey)
		ripemd160Hash := ripemd160.New()
		ripemd160Hash.Write(sha256Hash[:])
		hash := ripemd160Hash.Sum(nil)
		result := hex.EncodeToString(hash)

		return result
	}
	return
}

func GetBoundStateByAddress(api, address string) (bool, error) {
	apiurl, _ := url.JoinPath(api , valPass, address)
	response, err := http.Get(apiurl)
	if err != nil {
		fmt.Println("Ошибка при выполнении GET-запроса:", err)
		return false, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return false, err
	}

	var resp ValidatorResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("Ошибка при распаковке JSON:", err)
		return false, err
	}

	return resp.Validator.Jailed, nil
}

