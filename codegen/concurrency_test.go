package codegen

import (
	"strings"
	"testing"
)

func TestSpawnGeneratesAsyncFunction(t *testing.T) {
	input := "spawn task fetchData:\n    result is 42\n    give back result\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__abort_fetchData = new AbortController()") {
		t.Errorf("expected AbortController declaration, got:\n%s", output)
	}
	if !strings.Contains(output, "__task_fetchData = (async (signal)") {
		t.Errorf("expected async IIFE with signal, got:\n%s", output)
	}
}

func TestParallelGeneratesPromiseAll(t *testing.T) {
	input := "parallel:\n    task1 is fetch(\"/users\")\n    task2 is fetch(\"/posts\")\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Promise.all") {
		t.Errorf("expected Promise.all, got:\n%s", output)
	}
	if !strings.Contains(output, "task1, task2") {
		t.Errorf("expected destructured variable names, got:\n%s", output)
	}
}

func TestRaceGeneratesPromiseRace(t *testing.T) {
	input := "race:\n    fast is fetch(\"/cdn1\")\n    backup is fetch(\"/cdn2\")\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Promise.race") {
		t.Errorf("expected Promise.race, got:\n%s", output)
	}
}

func TestChannelGeneratesQuillChannel(t *testing.T) {
	input := "channel messages with buffer 10\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "__QuillChannel") {
		t.Errorf("expected __QuillChannel class, got:\n%s", output)
	}
	if !strings.Contains(output, "new __QuillChannel(10)") {
		t.Errorf("expected new __QuillChannel(10), got:\n%s", output)
	}
}

func TestChannelNoBufferGeneratesDefault(t *testing.T) {
	input := "channel events\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "new __QuillChannel()") {
		t.Errorf("expected new __QuillChannel(), got:\n%s", output)
	}
}

func TestSendGeneratesChannelSend(t *testing.T) {
	input := "channel messages\nsend \"hello\" to messages\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "await messages.send(") {
		t.Errorf("expected await channel.send(), got:\n%s", output)
	}
}

func TestReceiveGeneratesChannelReceive(t *testing.T) {
	input := "channel messages\nmsg is receive from messages\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "await messages.receive()") {
		t.Errorf("expected await channel.receive(), got:\n%s", output)
	}
}

func TestSelectGeneratesPromiseRaceWithChannels(t *testing.T) {
	input := `channel messages
channel errors
select:
    when receive from messages:
        say "got message"
    when receive from errors:
        say "got error"
    after 5000:
        say "timeout"
`
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Promise.race") {
		t.Errorf("expected Promise.race in select, got:\n%s", output)
	}
	if !strings.Contains(output, "messages.receive()") {
		t.Errorf("expected messages.receive() in select, got:\n%s", output)
	}
	if !strings.Contains(output, "errors.receive()") {
		t.Errorf("expected errors.receive() in select, got:\n%s", output)
	}
	if !strings.Contains(output, "setTimeout") {
		t.Errorf("expected setTimeout for after clause, got:\n%s", output)
	}
}

func TestAsyncIIFEWrapperWithSpawn(t *testing.T) {
	input := "spawn task myTask:\n    say \"hello\"\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "(async () => {") {
		t.Errorf("expected async IIFE wrapper, got:\n%s", output)
	}
	if !strings.Contains(output, "})();") {
		t.Errorf("expected async IIFE closing, got:\n%s", output)
	}
}

func TestChannelRuntimeInjection(t *testing.T) {
	input := "channel messages with buffer 5\n"
	output, err := compile(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "class __QuillChannel") {
		t.Errorf("expected __QuillChannel class injected, got:\n%s", output)
	}
	if !strings.Contains(output, "send(value)") {
		t.Errorf("expected send method in __QuillChannel, got:\n%s", output)
	}
	if !strings.Contains(output, "receive()") {
		t.Errorf("expected receive method in __QuillChannel, got:\n%s", output)
	}
}
