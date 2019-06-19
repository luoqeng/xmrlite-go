package main

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/luoqeng/mymonero-core-go/src"
)

const fromAddr = "9wq792k9sxVZiLn66S3Qzv8QfmtcwkdXgM5cWGsXAPxoQeMQ79md51PLPCijvzk1iHbuHi91pws5B7iajTX9KTtJ4bh2tCh"
const viewKey = "f747f4a4838027c9af80e6364a941b60c538e67e9ea198b6ec452b74c276de06"
const spendKey = "509a9761fde8856fc38e79ca705d85f979143524f178f8e2e0eb539fc050e905"
const pubSpendKey = "7b0ca4ee4cf32ac393b0306402d71b2c4a3db02c45ffbb780c2fde677b03848d"
const payID = "a32497eebed8aebb5d5a633d559f7eb7e819d2da183cafe41234567890abda48"
const sendAmount = "1000000000000"
const toAddr = "A2rgGdM78JEQcxEUsi761WbnJWsFRCwh1PkiGtGnUUcJTGenfCr5WEtdoXezutmPiQMsaM4zJbpdH5PMjkCt7QrXAhV8wDB"
const nettype = "TESTNET"

type UnspentOut struct {
	Amout       string `json:"amount"`
	PublicKey   string `json:"public_key"`
	Rct         string `json:"rct"`
	GlobalIndex string `json:"global_index"`
	Index       string `json:"index"`
	TxPubKey    string `json:"tx_pub_key"`
}

type MixOut struct {
	Amout   string       `json:"amount"`
	Outputs []UnspentOut `json:"outputs"`
}

type UnspentOuts struct {
	PassedInAttemptAtFee string       `json:"passedIn_attemptAt_fee"`
	PaymentIdString      string       `json:"payment_id_string"`
	SendingAmount        string       `json:"sending_amount"`
	IsSweeping           bool         `json:"is_sweeping"`
	Priority             string       `json:"priority"`
	FeePerB              string       `json:"fee_per_b"`
	FeeMask              string       `json:"fee_mask"`
	ForkVersion          string       `json:"fork_version"`
	UnspentOuts          []UnspentOut `json:"unspent_outs"`
}

type Decoys struct {
	Mixin           string       `json:"mixin"`
	UsingFee        string       `json:"using_fee"`
	FinalTotalWoFee string       `json:"final_total_wo_fee"`
	ChangeAmount    string       `json:"change_amount"`
	UsingOuts       []UnspentOut `json:"using_outs"`
}

type Transaction struct {
	PassedInAttemptAtFee string       `json:"passedIn_attemptAt_fee"`
	FromAddressString    string       `json:"from_address_string"`
	SecViewKeyString     string       `json:"sec_viewKey_string"`
	SecSpendKeyString    string       `json:"sec_spendKey_string"`
	ToAddressString      string       `json:"to_address_string"`
	PaymentIdString      string       `json:"payment_id_string"`
	FinalTotalWoFee      string       `json:"final_total_wo_fee"`
	ChangeAmount         string       `json:"change_amount"`
	FeeAmount            string       `json:"fee_amount"`
	Priority             string       `json:"priority"`
	FeePerB              string       `json:"fee_per_b"`
	FeeMask              string       `json:"fee_mask"`
	UnlockTime           string       `json:"unlock_time"`
	NettypeString        string       `json:"nettype_string"`
	ForkVersion          string       `json:"fork_version"`
	UsingOuts            []UnspentOut `json:"using_outs"`
	MixOuts              []MixOut     `json:"mix_outs"`
}

type SignedTx struct {
	TxMustBeReconstructed string `json:"tx_must_be_reconstructed"`
	FeeActuallyNeeded     string `json:"fee_actually_needed"`
	SerializedSignedTx    string `json:"serialized_signed_tx"`
	TxHash                string `json:"tx_hash"`
	TxPubKey              string `json:"tx_pub_key"`
	TxKey                 string `json:"tx_key"`
}

func main() {
	cli := NewClient("http://127.0.0.1:1984")
	resUnspentOuts, err := cli.GetUnspentOuts(fromAddr)
	if err != nil {
		log.Fatal(err)
	}
	unspentOuts, err := ParsedResGetUnspentOuts(resUnspentOuts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("unspent len: %d\n", len(unspentOuts.UnspentOuts))

	unspentOuts.PaymentIdString = payID
	unspentOuts.SendingAmount = sendAmount
	jsonStr, err := json.Marshal(unspentOuts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("json: %s\n", jsonStr)
	resDecoys := mymonero.CallFunc("send_step1__prepare_params_for_get_decoys", string(jsonStr))
	log.Printf("decoys: %s\n", resDecoys)
	errMsg, err := jsonparser.GetString([]byte(resDecoys), "err_msg")
	if err == nil {
		log.Fatal(errors.New(errMsg))
	}

	decoys, err := ParseGetDecoys([]byte(resDecoys))
	if err != nil {
		log.Fatal(err)
	}

	amounts, err := NewReqGetRandomOuts([]byte(resDecoys))
	if err != nil {
		log.Fatal(err)
	}
	resRandomOuts, err := cli.GetRandomOuts(amounts, 11)
	if err != nil {
		log.Fatal(err)
	}
	mixOuts, err := ParsedResGetRandomOuts(resRandomOuts)
	if err != nil {
		log.Fatal(err)
	}

	tx := Transaction{
		PassedInAttemptAtFee: unspentOuts.PassedInAttemptAtFee,
		FromAddressString:    fromAddr,
		SecViewKeyString:     viewKey,
		SecSpendKeyString:    spendKey,
		ToAddressString:      toAddr,
		PaymentIdString:      payID,
		FinalTotalWoFee:      decoys.FinalTotalWoFee,
		ChangeAmount:         decoys.ChangeAmount,
		FeeAmount:            decoys.UsingFee,
		Priority:             unspentOuts.Priority,
		FeePerB:              unspentOuts.FeePerB,
		FeeMask:              unspentOuts.FeeMask,
		UnlockTime:           "0",
		NettypeString:        nettype,
		ForkVersion:          unspentOuts.ForkVersion,
		UsingOuts:            decoys.UsingOuts,
		MixOuts:              mixOuts,
	}

	jsonStr, err = json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	resTx := mymonero.CallFunc("send_step2__try_create_transaction", string(jsonStr))
	errMsg, err = jsonparser.GetString([]byte(resTx), "err_msg")
	if err == nil {
		log.Fatal(errors.New(errMsg))
	}

	signedTx, err := ParseSignedTx([]byte(resTx))
	if err != nil {
		log.Fatal(err)
	}

	signedTxStr, _ := json.Marshal(signedTx)
	log.Printf("%s\n", signedTxStr)

	if signedTx.TxMustBeReconstructed == "true" {
		log.Fatalln("PassedInAttemptAtFee = tx.FeeActuallyNeeded; // -> reconstruction attempt's step1's PassedInAttemptAtFee")
	}

	res, err := cli.SubmitRawTx(signedTx.SerializedSignedTx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s\n", res)
}

func ParsedResGetUnspentOuts(res []byte) (*UnspentOuts, error) {
	var (
		unspent UnspentOuts
		err     error
	)
	jsonparser.ArrayEach(res,
		func(value []byte,
			dataType jsonparser.ValueType,
			offset int, dataErr error) {
			if dataErr != nil {
				return
			}

			var output UnspentOut
			output.TxPubKey, err = jsonparser.GetString(value, "tx_pub_key")
			if err != nil {
				return
			}
			index, err := jsonparser.GetInt(value, "index")
			if err != nil {
				return
			}
			output.Index = strconv.FormatInt(index, 10)

			keyImage, err := GenerateKeyImage(pubSpendKey, spendKey, viewKey, output.TxPubKey, output.Index)
			if err != nil {
				return
			}

			isOutputSpent := false
			jsonparser.ArrayEach(value,
				func(spendKeyImage []byte,
					dataType jsonparser.ValueType,
					offset int, dataErr error) {
					if dataErr != nil {
						return
					}
					if string(spendKeyImage) == keyImage {
						isOutputSpent = true
					}
				}, "spend_key_images")
			// 过滤已经使用的输出
			if isOutputSpent == true {
				return
			}

			output.Amout, err = jsonparser.GetString(value, "amount")
			if err != nil {
				return
			}
			output.PublicKey, err = jsonparser.GetString(value, "public_key")
			if err != nil {
				return
			}
			output.Rct, err = jsonparser.GetString(value, "rct")
			if err != nil {
				return
			}

			globalIndex, err := jsonparser.GetInt(value, "global_index")
			if err != nil {
				return
			}
			output.GlobalIndex = strconv.FormatInt(globalIndex, 10)

			unspent.UnspentOuts = append(unspent.UnspentOuts, output)
		}, "outputs")

	if err != nil {
		return nil, err
	}
	fee, err := jsonparser.GetInt(res, "per_byte_fee")
	if err != nil {
		return nil, err
	}
	unspent.FeePerB = strconv.FormatInt(fee, 10)
	unspent.FeeMask = "10000"
	unspent.PassedInAttemptAtFee = "0"
	unspent.IsSweeping = false
	unspent.Priority = "1"
	unspent.ForkVersion = "0"

	return &unspent, nil
}

func ParseGetDecoys(res []byte) (*Decoys, error) {
	var (
		decoys Decoys
		err    error
	)

	err = json.Unmarshal(res, &decoys)
	return &decoys, err
	/*
		decoys.Mixin, err = jsonparser.GetString(res, "mixin")
		if err != nil {
			return nil, err
		}
		decoys.UsingFee, err = jsonparser.GetString(res, "using_fee")
		if err != nil {
			return nil, err
		}
		decoys.FinalTotalWoFee, err = jsonparser.GetString(res, "final_total_wo_fee")
		if err != nil {
			return nil, err
		}
		decoys.ChangeAmount, err = jsonparser.GetString(res, "change_amount")
		if err != nil {
			return nil, err
		}

		jsonparser.ArrayEach(res,
			func(value []byte,
				dataType jsonparser.ValueType,
				offset int, dataErr error) {
				if dataErr != nil {
					return
				}

				var output UnspentOut
				output.Amout, err = jsonparser.GetString(value, "amount")
				if err != nil {
					return
				}
				output.PublicKey, err = jsonparser.GetString(value, "public_key")
				if err != nil {
					return
				}
				output.Rct, err = jsonparser.GetString(value, "rct")
				if err != nil {
					return
				}
				output.GlobalIndex, err = jsonparser.GetString(value, "global_index")
				if err != nil {
					return
				}
				output.Index, err = jsonparser.GetString(value, "index")
				if err != nil {
					return
				}
				output.TxPubKey, err = jsonparser.GetString(value, "tx_pub_key")
				if err != nil {
					return
				}
				decoys.UsingOuts = append(decoys.UsingOuts, output)

			}, "using_outs")

		return &decoys, err
	*/
}

func NewReqGetRandomOuts(res []byte) (amounts []string, err error) {
	jsonparser.ArrayEach(res,
		func(value []byte,
			dataType jsonparser.ValueType,
			offset int, dataErr error) {
			if dataErr != nil {
				return
			}

			var (
				rct    string
				amount string
			)
			amount, err = jsonparser.GetString(value, "amount")
			if err != nil {
				return
			}
			rct, _ = jsonparser.GetString(value, "rct")
			if rct != "" {
				amounts = append(amounts, "0")
			} else {
				amounts = append(amounts, amount)
			}
		}, "using_outs")

	return amounts, err
}

func ParsedResGetRandomOuts(res []byte) (mixOuts []MixOut, err error) {
	jsonparser.ArrayEach(res,
		func(value []byte,
			dataType jsonparser.ValueType,
			offset int, dataErr error) {
			if dataErr != nil {
				return
			}

			var mixOut MixOut
			mixOut.Amout, err = jsonparser.GetString(value, "amount")
			if err != nil {
				return
			}

			jsonparser.ArrayEach(value,
				func(v []byte,
					dataType jsonparser.ValueType,
					offset int, dataErr error) {
					if dataErr != nil {
						return
					}

					var output UnspentOut
					output.PublicKey, err = jsonparser.GetString(v, "public_key")
					if err != nil {
						return
					}
					output.Rct, err = jsonparser.GetString(v, "rct")
					if err != nil {
						return
					}
					globalIndex, err := jsonparser.GetInt(v, "global_index")
					if err != nil {
						return
					}
					output.GlobalIndex = strconv.FormatInt(globalIndex, 10)

					mixOut.Outputs = append(mixOut.Outputs, output)
				}, "outputs")

			mixOuts = append(mixOuts, mixOut)
		}, "amount_outs")

	return mixOuts, err
}

func ParseSignedTx(res []byte) (*SignedTx, error) {
	var (
		tx  SignedTx
		err error
	)
	tx.TxMustBeReconstructed, err = jsonparser.GetString(res, "tx_must_be_reconstructed")
	if err != nil {
		return nil, err
	}
	if tx.TxMustBeReconstructed == "true" {
		tx.FeeActuallyNeeded, err = jsonparser.GetString(res, "fee_actually_needed")
		if err != nil {
			return nil, err
		}
		return &tx, nil
	}

	tx.SerializedSignedTx, err = jsonparser.GetString(res, "serialized_signed_tx")
	if err != nil {
		return nil, err
	}
	tx.TxHash, err = jsonparser.GetString(res, "tx_hash")
	if err != nil {
		return nil, err
	}
	tx.TxPubKey, err = jsonparser.GetString(res, "tx_pub_key")
	if err != nil {
		return nil, err
	}
	tx.TxKey, err = jsonparser.GetString(res, "tx_key")
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func GenerateKeyImage(pubSpendKey, secSpendKey, secViewKey, txPubKey, outIndex string) (string, error) {
	args := make(map[string]string)
	args["pub_spendKey_string"] = pubSpendKey
	args["sec_spendKey_string"] = secSpendKey
	args["sec_viewKey_string"] = secViewKey
	args["tx_pub_key"] = txPubKey
	args["out_index"] = outIndex

	argsStr, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	ret := mymonero.CallFunc("generate_key_image", string(argsStr))

	errMsg, err := jsonparser.GetString([]byte(ret), "err_msg")
	if err == nil {
		return "", errors.New(errMsg)
	}

	keyImage, err := jsonparser.GetString([]byte(ret), "retVal")
	if err != nil {
		return "", err
	}
	return keyImage, nil
}
