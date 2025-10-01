# 🃏 Scrum Planning Poker CLI Tool (Go)

A lightweight **Scrum Planning Poker server** built in **Go**, using **CLI-only** interface. This tool allows team members to create or join rooms and vote using Fibonacci-style estimates in real-time.


https://github.com/user-attachments/assets/6ba56e2f-9e8e-4542-8267-0be53ad75306


---

## 🚀 Features

- ✅ Create or join voting rooms using a UUID
- 👥 Support for multiple simultaneous users via TCP connections
- 🔢 Fibonacci voting (0, 1, 2, 3, 5, 8, ...)
- 🕵️ Shows **missing voters** during active sessions
- 📊 Displays results once all participants have voted
- 🔄 Resets votes after each round for repeated use

---

## 🛠️ Getting Started
1. **Start the Server**
```bash
go run ./server
```
This command starts the Planning Poker server on your local machine, listening for TCP connections on port 9000. Clients can connect to this server to create or join voting rooms.

2. **Start the Client**
```bash
go run ./client -host=<hostname>
```
This command starts the client and connects it to the specified server host. Replace <hostname> with the address of the machine running the server (e.g., localhost or a remote IP).


3. **Start the App Using the Binaries**  
You can use the pre-built binaries to run the application.  
Download the appropriate binary for your operating system from the [Releases page](https://github.com/Mbauro/party-goker/releases).

Choose the one based on your operating system, then run it:

**For example, on macOS:**

To start the **client**:

```bash
./party-goker-client-darwin-amd64 --host=<hostname>
```

To start the **server**:

```bash
./party-goker-srv-darwin-amd64
```

4. **Compile the binaries Versions**
You can build executable binaries for both the server and client:

```bash
go build -o party-gokker-srv ./server
go build -o party-gokker ./client
```
These commands generate the following binaries:

*party-gokker-srv* — the server

*party-gokker* — the client

To run them:

```bash
./party-gokker-srv
```
or

```bash
./party-gokker -host=<hostname>
```
This is useful for deployment, sharing, or running without needing the Go toolchain installed.
