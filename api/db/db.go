// Package db provides a database connection for the application. SQLite is used as the database engine.
package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/ncruces/go-sqlite3/driver"
)

/**
 * Database tables
 - documents: stores the documents to be summarized
	+ uuid: unique identifier for each document
	+ hash: hash of the document content for deduplication
	+ content: content of the document
	+ summary: summary of the document
	+ created_at: timestamp of when the document was created
 - chats: stores the chat history between the user and the LLM
	+ uuid: unique identifier for each chat
	+ document_uuid: foreign key to the document table
	+ user_message: user's message
	+ llm_message: LLM's message
	+ created_at: timestamp of when the chat was created
*/

type Document struct {
	UUID      string `db:"uuid"`
	Hash      string `db:"hash"`
	Content   string `db:"content"`
	Summary   string `db:"summary"`
	CreatedAt string `db:"created_at"`
}

type Chat struct {
	UUID         string `db:"uuid"`
	DocumentUUID string `db:"document_uuid"`
	UserMessage  string `db:"user_message"`
	LLMMessage   string `db:"llm_message"`
	CreatedAt    string `db:"created_at"`
}

// DB is a struct representing a database connection.
type DB struct {
	db *sql.DB
}

// init database tables if they don't exist
func (db *DB) Init() error {
	_, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS documents (
		uuid TEXT PRIMARY KEY,
		hash TEXT NOT NULL,
		content TEXT NOT NULL,
		summary TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	_, err = db.db.Exec(`CREATE TABLE IF NOT EXISTS chats (
		uuid TEXT PRIMARY KEY,
		document_uuid TEXT NOT NULL,
		user_message TEXT NOT NULL,
		llm_message TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	return nil
}

func newDBConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "file:db/db.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	// Check SQLite version
	var version string
	err = db.QueryRow(`SELECT sqlite_version()`).Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query SQLite version: %v", err)
	}
	log.Printf("Using SQLite version: %s", version)

	return db
}

// NewDB creates a new database connection.
func NewDB() *DB {
	return &DB{db: newDBConnection()}
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.db.Close()
}

func generateUUID() string {
	return uuid.New().String()
}

func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// CreateDocument creates a new document in the database.
func (db *DB) CreateDocument(hash, content, summary string) (string, error) {
	stmt, err := db.db.Prepare("INSERT INTO documents (uuid, hash, content, summary, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	documentUUID := generateUUID()
	_, err = stmt.Exec(documentUUID, hash, content, summary, getCurrentTimestamp())
	if err != nil {
		return "", err
	}

	return documentUUID, nil
}

// GetDocuments returns a list of documents from the database.
func (db *DB) GetDocuments() ([]Document, error) {
	rows, err := db.db.Query("SELECT * FROM documents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var document Document
		err := rows.Scan(&document.UUID, &document.Hash, &document.Content, &document.Summary, &document.CreatedAt)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	return documents, nil
}

// GetDocumentByID returns a document from the database by its ID.
func (db *DB) GetDocumentByID(id string) (Document, error) {
	var document Document
	err := db.db.QueryRow("SELECT * FROM documents WHERE uuid = ?", id).Scan(&document.UUID, &document.Hash, &document.Content, &document.Summary, &document.CreatedAt)
	if err != nil {
		return Document{}, err
	}
	return document, nil
}

// GetDocumentByHash returns a document from the database by its hash.
func (db *DB) GetDocumentByHash(hash string) (Document, error) {
	var document Document
	err := db.db.QueryRow("SELECT * FROM documents WHERE hash = ?", hash).Scan(&document.UUID, &document.Hash, &document.Content, &document.Summary, &document.CreatedAt)
	if err != nil {
		return Document{}, err
	}
	return document, nil
}

// CreateChat creates a new chat in the database.
func (db *DB) CreateChat(documentUUID, userMessage, llmMessage string) (string, error) {
	stmt, err := db.db.Prepare("INSERT INTO chats (uuid, document_uuid, user_message, llm_message, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	chatUUID := generateUUID()
	_, err = stmt.Exec(chatUUID, documentUUID, userMessage, llmMessage, getCurrentTimestamp())
	if err != nil {
		return "", err
	}

	return chatUUID, nil
}

// GetChats returns a list of chats from the database.
func (db *DB) GetChats() ([]Chat, error) {
	rows, err := db.db.Query("SELECT * FROM chats")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []Chat
	for rows.Next() {
		var chat Chat
		err := rows.Scan(&chat.UUID, &chat.DocumentUUID, &chat.UserMessage, &chat.LLMMessage, &chat.CreatedAt)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

// GetChatByID returns a chat from the database by its ID.
func (db *DB) GetChatByID(id string) (Chat, error) {
	var chat Chat
	err := db.db.QueryRow("SELECT * FROM chats WHERE uuid = ?", id).Scan(&chat.UUID, &chat.DocumentUUID, &chat.UserMessage, &chat.LLMMessage, &chat.CreatedAt)
	if err != nil {
		return Chat{}, err
	}
	return chat, nil
}
