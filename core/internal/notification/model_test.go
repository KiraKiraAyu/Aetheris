package notification

import (
	"database/sql/driver"
	"reflect"
	"testing"
)

func TestMetadataValueAndScanRoundTrip(t *testing.T) {
	metadata := Metadata{"invoice_id": "inv_123", "source": "billing"}

	value, err := metadata.Value()
	if err != nil {
		t.Fatalf("Value returned error: %v", err)
	}
	if _, ok := value.(driver.Value); !ok {
		t.Fatalf("Value returned non-driver value: %T", value)
	}

	var scanned Metadata
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if !reflect.DeepEqual(scanned, metadata) {
		t.Fatalf("scanned metadata = %#v, want %#v", scanned, metadata)
	}
}

func TestMetadataScanHandlesNilAndUnsupportedTypes(t *testing.T) {
	var nilMetadata Metadata
	if err := nilMetadata.Scan(nil); err != nil {
		t.Fatalf("Scan nil returned error: %v", err)
	}
	if len(nilMetadata) != 0 {
		t.Fatalf("nil scan = %#v, want empty metadata", nilMetadata)
	}

	var bad Metadata
	if err := bad.Scan(123); err == nil {
		t.Fatal("Scan unsupported type should return an error")
	}
}

func TestMetadataCloneIsIndependent(t *testing.T) {
	metadata := Metadata{"source": "billing"}
	clone := metadata.Clone()
	clone["source"] = "crm"

	if metadata["source"] != "billing" {
		t.Fatalf("original metadata changed to %q", metadata["source"])
	}
}

func TestChannelValidIncludesCommonChatReceivers(t *testing.T) {
	for _, channel := range []Channel{
		ChannelTelegram,
		ChannelSlack,
		ChannelDiscord,
		ChannelFeishu,
		ChannelDingTalk,
		ChannelWeCom,
	} {
		if !channel.Valid() {
			t.Fatalf("channel %q should be valid", channel)
		}
	}
}
