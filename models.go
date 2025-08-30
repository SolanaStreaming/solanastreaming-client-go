package solanastreaming

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

var (
	ErrConnectFirst       = errors.New("connect first")
	ErrNoSubscription     = errors.New("no subscription")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
	ErrSubscriptionClosed = errors.New("subscription closed")
)

type receiverType int

const (
	receiverTypeByRequestID receiverType = iota
	receiverTypeBySubscriptionID
)

type receiver struct {
	Type  receiverType
	Value int
}

type wireMessage struct {
	ID             int              `json:"id"`
	SubscriptionID uint             `json:"subscription_id,omitempty"` // for reeiving subscription notifications
	Method         string           `json:"method"`                    // method or notification name
	Params         *json.RawMessage `json:"params,omitempty"`          // notification body or request params
	Result         json.RawMessage  `json:"result"`                    // response to subscription messages
	Error          *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type LatestBlockNotification struct {
	Block     uint64 `json:"block"`
	BlockTime uint64 `json:"blockTime"`
}

type NewPairNotification struct {
	Slot      uint64 `json:"slot"`
	Signature string `json:"signature"`
	BlockTime uint64 `json:"blockTime"`
	Pair      Pair   `json:"pair"`
}

type SwapNotification struct {
	Slot      uint64 `json:"slot"`
	Signature string `json:"signature"`
	BlockTime uint64 `json:"blockTime"`
	Swap      Swap   `json:"swap"`
}

type Pair struct {
	SourceExchange           string           `json:"sourceExchange"`           // The exchange where the pair is listed
	AmmAccount               solana.PublicKey `json:"ammAccount"`               // The address of the AMM account for the pair, or the bonding curve for launch platforms.
	BaseToken                Token            `json:"baseToken"`                // The base token in the pair
	QuoteToken               Token            `json:"quoteToken"`               // The quote token in the pair, usually wrapped SOL
	BaseTokenLiquidityAdded  string           `json:"baseTokenLiquidityAdded"`  // The amount of base token added to the liquidity pool
	QuoteTokenLiquidityAdded string           `json:"quoteTokenLiquidityAdded"` // The amount of quote token added to the liquidity pool
	Migration                string           `json:"migration,omitempty"`      // If this pair was the result of a launch migration, this field will contain the migration source exchange. e.g. "pumpfun" or "raydium_launchpad"
}

type Token struct {
	Account solana.PublicKey `json:"account"` // The address of the token account (the token mint address)
	Info    *TokenInfo       `json:"info"`    // Additional token information, including metadata and authority
}

type TokenInfo struct {
	Decimals        uint              `json:"decimals"`        // The number of decimal places for the token.
	Supply          string            `json:"supply"`          // The total supply of the token, represented as a string to handle large numbers.
	MetaData        *TokenMetaData    `json:"metadata"`        // Metadata about the token, such as name, symbol, and logo.
	MintAuthority   *solana.PublicKey `json:"mintAuthority"`   // The authority that can mint new tokens. (nil for non-mintable tokens)
	FreezeAuthority *solana.PublicKey `json:"freezeAuthority"` // The authority that can freeze token accounts. (nil for non-freezable tokens)
}

type TokenMetaData struct {
	Name    string       `json:"name"`
	Symbol  string       `json:"symbol"`
	Logo    string       `json:"logo"`
	Socials TokenSocials `json:"socials"`
}
type TokenSocials struct {
	Website  string `json:"website"`
	X        string `json:"x"`
	Telegram string `json:"telegram"`
}

type Swap struct {
	SourceExchange      string           `json:"sourceExchange"`      // The source exchange of the swap
	AmmAccount          solana.PublicKey `json:"ammAccount"`          // The same as dexscreener pair address. This can also be the address of the bondin curve in launch token swaps like pump.fun and raydium launchpad.
	BaseTokenMint       solana.PublicKey `json:"baseTokenMint"`       // The token being traded
	QuoteTokenMint      solana.PublicKey `json:"quoteTokenMint"`      // The token being traded for and what the price is quoted in. Note: this is usually wrapped sol.
	WalletAccount       solana.PublicKey `json:"walletAccount"`       // The wallet in the swap thats not the pool.
	QuotePrice          *big.Float       `json:"quotePrice"`          // The execution price of the swap. This field is calculated as the quoteAmount divided by the baseAmount.
	USDValue            float64          `json:"usdValue"`            // The value of the swap in USD.
	BaseAmount          string           `json:"baseAmount"`          // The amount of base token traded in the swap
	SwapType            string           `json:"swapType"`            // The type of swap (buy/sell)
	QuoteTokenLiquidity string           `json:"quoteTokenLiquidity"` // (Beta Testing) The amount of quote token in the liquidity pool. This is not always available and could be an empty string.
}
