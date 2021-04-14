package eyeshade

import (
	"github.com/brave-intl/bat-go/eyeshade/avro"
	avrocontribution "github.com/brave-intl/bat-go/eyeshade/avro/contribution"
	avroreferral "github.com/brave-intl/bat-go/eyeshade/avro/referral"
	avrosettlement "github.com/brave-intl/bat-go/eyeshade/avro/settlement"
	avrosuggestion "github.com/brave-intl/bat-go/eyeshade/avro/suggestion"
	"github.com/segmentio/kafka-go"
)

var (
	// Handlers is a map for a topic key to point to any non standard handlers
	// all others are handled by HandlerDefault
	Handlers = map[string]func(con *MessageHandler, msgs []kafka.Message) error{
		"suggestion":   HandleVotes,
		"contribution": HandleVotes,
		"settlement":   HandlerInsertConvertableTransaction,
		"referral":     HandlerInsertConvertableTransaction,
	}
	// DecodeBatchVotes a mapping to help the batch decoder find it's topic specific decoder
	DecodeBatchVotes = map[string]avro.BatchVoteDecoder{
		"suggestion":   avrosuggestion.DecodeBatch,
		"contribution": avrocontribution.DecodeBatch,
	}
	// DecodeBatchTransactions a mapping to help the batch decoder find it's topic specific decoder
	DecodeBatchTransactions = map[string]avro.BatchConvertableTransactionDecoder{
		"referral":   avroreferral.DecodeBatch,
		"settlement": avrosettlement.DecodeBatch,
	}
)

// HandleVotes handles vote insertions
func HandleVotes(
	con *MessageHandler,
	msgs []kafka.Message,
) error {
	votes, err := DecodeBatchVotes[con.key](
		KeyToEncoder[con.key].Codecs(),
		msgs,
	)
	if err != nil {
		return err
	}
	return con.service.Datastore(false).
		InsertVotes(con.Context(), *votes)
}

// HandlerInsertConvertableTransaction is the default handler for direct to transaction use cases
func HandlerInsertConvertableTransaction(
	con *MessageHandler,
	msgs []kafka.Message,
) error {
	modifiers, err := con.Modifiers()
	if err != nil {
		return err
	}
	txs, err := DecodeBatchTransactions[con.key](
		KeyToEncoder[con.key].Codecs(),
		msgs,
		modifiers...,
	)
	if err != nil {
		return err
	}
	return con.service.InsertConvertableTransactions(
		con.Context(),
		*txs,
	)
}
