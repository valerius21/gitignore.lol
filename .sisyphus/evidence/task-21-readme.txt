TASK 21: Fix README.md Inaccuracies
====================================

VERIFICATION RESULTS:
✓ Go version: Changed from 1.21.6 to 1.25 (lines 40, 75)
✓ Build command: Changed from `go build -o gitignore-server ./cmd/gitignore_server.go` to `go build -o gitignore-lol ./cmd/main.go` (line 52)
✓ Binary name: Changed from `./gitignore-server` to `./gitignore-lol` (lines 55, 124, 127)
✓ Port default: Changed from 3000 to 4444 (line 109)
✓ Bun prerequisite: Added to Development > Prerequisites section (line 76)
✓ Script reference: Changed from `./scripts/generate-swagger.sh` to `./scripts/generate_docs.sh` (line 104)
✓ CLI flags: Verified against pkg/lib/cli.go - all flags match (line 138 shows --port=4444)

GREP OUTPUT (verification):
40:- Go 1.25 or later
52:go build -o gitignore-lol ./cmd/main.go
55:./gitignore-lol
75:- Go 1.25 or later
104:./scripts/generate_docs.sh
109:- `PORT` - Server port (default: 4444)
124:./gitignore-lol --rate-limit 50 --rate-window 30
127:./gitignore-lol --enable-rate-limit=false
138:      --port=4444                                         Port the server listens on.

NO INACCURACIES REMAINING:
- No references to 1.21.6
- No references to gitignore-server binary
- No references to 3000 port default
- No references to non-existent generate-swagger.sh script
