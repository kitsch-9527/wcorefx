//go:build windows

package event

import (
	"testing"
)

func TestUnmarshalXML(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8"?>
<Event xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
  <System>
    <Provider Name="TestProvider" Guid="{00000000-0000-0000-0000-000000000000}"/>
    <EventID>1000</EventID>
    <Version>0</Version>
    <Level>4</Level>
    <Task>0</Task>
    <Opcode>0</Opcode>
    <Keywords>9223372036854775808</Keywords>
    <TimeCreated SystemTime="2024-01-15T10:30:00.000000000Z"/>
    <EventRecordID>12345</EventRecordID>
    <Channel>Application</Channel>
    <Computer>TestPC</Computer>
    <Security UserID="S-1-5-18"/>
  </System>
  <EventData>
    <Data Name="Param1">Value1</Data>
  </EventData>
  <RenderingInfo>
    <Message>Test message</Message>
    <Level>Information</Level>
    <Task>None</Task>
    <Opcode>Info</Opcode>
    <Keywords>
      <Keyword>Classic</Keyword>
    </Keywords>
  </RenderingInfo>
</Event>`

	event, err := UnmarshalXML([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}

	if event.EventIdentifier.ID != 1000 {
		t.Errorf("EventID = %d, want 1000", event.EventIdentifier.ID)
	}
	if event.Provider.Name != "TestProvider" {
		t.Errorf("Provider.Name = %q, want TestProvider", event.Provider.Name)
	}
	if event.LevelRaw != 4 {
		t.Errorf("LevelRaw = %d, want 4", event.LevelRaw)
	}
	if event.Channel != "Application" {
		t.Errorf("Channel = %q, want Application", event.Channel)
	}
	if event.Computer != "TestPC" {
		t.Errorf("Computer = %q, want TestPC", event.Computer)
	}
	if event.RecordID != 12345 {
		t.Errorf("RecordID = %d, want 12345", event.RecordID)
	}
	if event.Message != "Test message" {
		t.Errorf("Message = %q, want Test message", event.Message)
	}
	if event.Level != "Information" {
		t.Errorf("Level = %q, want Information", event.Level)
	}
	if len(event.Keywords) != 1 || event.Keywords[0] != "Classic" {
		t.Errorf("Keywords = %v, want [Classic]", event.Keywords)
	}
	if event.User.Identifier != "S-1-5-18" {
		t.Errorf("User.Identifier = %q, want S-1-5-18", event.User.Identifier)
	}

	// Check TimeCreated
	tm := event.TimeCreated.SystemTime
	if tm.Year() != 2024 || tm.Month() != 1 || tm.Day() != 15 {
		t.Errorf("TimeCreated = %v, want 2024-01-15", tm)
	}
}

func TestUnmarshalXMLInvalid(t *testing.T) {
	_, err := UnmarshalXML([]byte("invalid xml"))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestEnrichRawValuesWithNames(t *testing.T) {
	ev := &Event{
		KeywordsRaw: 0x10000000000000,
		LevelRaw:    2,
		TaskRaw:     0,
	}
	taskRaw := uint8(0)
	ev.OpcodeRaw = &taskRaw

	EnrichRawValuesWithNames(nil, ev)

	if len(ev.Keywords) != 1 || ev.Keywords[0] != "Audit Failure" {
		t.Errorf("Keywords = %v, want [Audit Failure]", ev.Keywords)
	}
	if ev.Level != "Error" {
		t.Errorf("Level = %q, want Error", ev.Level)
	}
	if ev.Task != "None" {
		t.Errorf("Task = %q, want None", ev.Task)
	}
	if ev.Opcode != "Info" {
		t.Errorf("Opcode = %q, want Info", ev.Opcode)
	}
}

func TestEnrichRawValuesWithNamesPublisherMeta(t *testing.T) {
	// Use values NOT in defaultWinMeta to trigger publisher lookup
	publisherMeta := &WinMeta{
		Keywords: map[int64]string{
			0x2: "CustomKeyword",
		},
		Levels: map[uint8]string{
			6: "CustomLevel", // 6 is not in defaultWinMeta (0-5)
		},
		Tasks: map[uint16]string{
			99: "CustomTask", // 99 is not in defaultWinMeta (only 0)
		},
	}

	ev := &Event{
		KeywordsRaw: 0x2,
		LevelRaw:    6,
		TaskRaw:     99,
	}

	EnrichRawValuesWithNames(publisherMeta, ev)

	if len(ev.Keywords) != 1 || ev.Keywords[0] != "CustomKeyword" {
		t.Errorf("Keywords = %v, want [CustomKeyword]", ev.Keywords)
	}
	if ev.Level != "CustomLevel" {
		t.Errorf("Level = %q, want CustomLevel", ev.Level)
	}
	if ev.Task != "CustomTask" {
		t.Errorf("Task = %q, want CustomTask", ev.Task)
	}
}

func TestFormat(t *testing.T) {
	records := []Record{
		{
			Event: Event{
				EventIdentifier: EventIdentifier{ID: 100},
				Provider:        Provider{Name: "Test"},
				Level:           "Error",
			},
		},
	}
	results := Format(records, nil)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].EventID != 100 {
		t.Errorf("EventID = %d, want 100", results[0].EventID)
	}
}
