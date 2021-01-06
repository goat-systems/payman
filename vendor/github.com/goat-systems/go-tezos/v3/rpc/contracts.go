package rpc

import (
	"encoding/hex"
	"fmt"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/goat-systems/go-tezos/v3/internal/crypto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"
)

/*
ScriptExpression is a string that will eventually be forged into a script_expression
See: https://tezos.gitlab.io/api/rpc.html#get-block-id-context-big-maps-big-map-id-script-expr
*/
type ScriptExpression string

/*
BigMapInput is the input for the goTezos.BigMap function.

Function:
	func (t *GoTezos) BigMap(input BigMapInput) ([]byte, error) {}
*/
type BigMapInput struct {
	Cycle            int
	Blockhash        string
	BigMapID         int              `validate:"required"`
	ScriptExpression ScriptExpression `validate:"required"`
}

func (b *BigMapInput) validate() error {
	if b.Blockhash == "" && b.Cycle == 0 {
		return errors.New("invalid input: missing key cycle or blockhash")
	} else if b.Blockhash != "" && b.Cycle != 0 {
		return errors.New("invalid input: cannot have both cycle and blockhash")
	}

	err := validator.New().Struct(b)
	if err != nil {
		return errors.Wrap(err, "invalid input")
	}

	return nil
}

/*
ContractStorage gets access the data of the contract.

Path:
	../<block_id>/context/contracts/<contract_id>/storage (GET)

Link:
	https://tezos.gitlab.io/api/rpc.html#get-block-id-context-contracts-contract-id-storage

Parameters:

	blockhash:
		The hash of block (height) of which you want to make the query.

	KT1:
		The contract address.
*/
func (c *Client) ContractStorage(blockhash string, KT1 string) ([]byte, error) {
	query := fmt.Sprintf("/chains/main/blocks/%s/context/contracts/%s/storage", blockhash, KT1)
	resp, err := c.get(query)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not get storage '%s'")
	}

	return resp, nil
}

/*
ContractStorage gets access the data of the contract.

Path:
 	../<block_id>/context/big_maps/<big_map_id>/<script_expr> (GET)

Link:
	https://tezos.gitlab.io/api/rpc.html#get-block-id-context-big-maps-big-map-id-script-expr

Parameters:

	bigMapID:
		The ID of the big_map

	scriptExpression:
		The scriptExpression of the big_map to query


*/
func (c *Client) BigMap(input BigMapInput) ([]byte, error) {
	if input.Cycle != 0 {
		snapshot, err := c.Cycle(input.Cycle)
		if err != nil {
			return []byte{}, errors.Wrapf(err, "could not get big map '%d' at cycle '%d'", input.BigMapID, input.Cycle)
		}

		input.Blockhash = snapshot.BlockHash
	}

	query := fmt.Sprintf("/chains/main/blocks/%s/context/big_maps/%d/%s", input.Blockhash, input.BigMapID, input.ScriptExpression)
	resp, err := c.get(query)
	if err != nil {
		return []byte{}, errors.Wrap(err, "could not get storage '%s'")
	}

	return resp, nil
}

/*
ForgeScriptExpressionForAddress -
A helper to forge a source account for a big_map script_exp
*/
func ForgeScriptExpressionForAddress(input string) (ScriptExpression, error) {
	input, err := pack(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to forge script expression for address")
	}

	prefix := []byte{13, 44, 64, 27}

	a := []byte{}
	for i := 0; i < len(input); i += 2 {
		elem, err := hex.DecodeString(input[i:(i + 2)])
		if err != nil {
			return "", errors.Wrap(err, "failed to forge script expression for address")
		}
		a = append(a, elem...)
	}

	hash, err := blake2b.New(32, []byte{})
	if err != nil {
		return "", errors.Wrap(err, "failed to forge script expression for address")
	}

	_, err = hash.Write(a)
	if err != nil {
		return "", errors.Wrap(err, "failed to forge script expression for address")
	}

	blakeHash := hash.Sum([]byte{})
	n := []byte{}
	n = append(n, prefix...)
	n = append(n, blakeHash...)

	return ScriptExpression(crypto.Encode(n)), nil
}

func pack(input string) (string, error) {
	var prefix []byte
	var tzPrefix string
	if strings.HasPrefix(input, "tz1") {
		prefix = []byte{6, 161, 159}
		tzPrefix = "00"
	} else if strings.HasPrefix(input, "tz2") {
		prefix = []byte{6, 161, 161}
		tzPrefix = "01"
	} else if strings.HasPrefix(input, "tz3") {
		prefix = []byte{6, 161, 164}
		tzPrefix = "02"
	}

	dInput, err := crypto.Decode(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to pack script expression")
	}
	bytes := fmt.Sprintf("00%s%s", tzPrefix, hex.EncodeToString(dInput[len(prefix):]))
	bytesHalfLen := len(bytes) / 2

	out := "050a"

	x := fmt.Sprintf("%x", bytesHalfLen)
	for len(x) < 8 {
		x = fmt.Sprintf("0%s", x)
	}

	out = fmt.Sprintf("%s%s", out, x)

	return fmt.Sprintf("%s%s", out, bytes), nil
}
