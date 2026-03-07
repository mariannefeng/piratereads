import { highlight } from "fumadocs-core/highlight";
import { CodeBlock } from "fumadocs-ui/components/codeblock";
import { ApiDemo } from "./api-demo";

const CODE = `import { useState } from "react";

function parseUserId(input) {
  const trimmed = input.trim();
  if (/^\d+$/.test(trimmed)) return trimmed;
  const match = trimmed.match(/goodreads\.com\/user\/show\/(\d+)/);
  return match ? match[1] : null;
}

export function GoodreadsShelf() {
  const [input, setInput] = useState("");
  const [books, setBooks] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  async function fetchShelf() {
    const userId = parseUserId(input);
    if (!userId) {
      setError("Paste a Goodreads profile URL or numeric user ID.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(
        \`https://api.piratereads.com/\${userId}/read?per_page=5\`,
      );
      if (!res.ok) throw new Error(\`\${res.status} \${res.statusText}\`);
      const data = await res.json();
      setBooks(data.books);
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <div style={{ display: "flex", gap: 8 }}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && fetchShelf()}
          placeholder="https://www.goodreads.com/user/show/12345"
          style={{
            flex: 1,
            padding: "6px 10px",
            borderRadius: 6,
            border: "1px solid #ccc",
          }}
        />
        <button
          onClick={fetchShelf}
          disabled={loading || !input.trim()}
          style={{ padding: "6px 14px", borderRadius: 6 }}
        >
          {loading ? "Loading…" : "Fetch shelf"}
        </button>
      </div>

      {error && <p style={{ color: "red" }}>{error}</p>}

      {books?.map((book) => (
        <a
          key={book.book_link}
          href={book.book_link}
          target="_blank"
          rel="noreferrer"
          style={{
            display: "flex",
            gap: 12,
            padding: 10,
            border: "1px solid #eee",
            borderRadius: 8,
            textDecoration: "none",
          }}
        >
          {book.book_cover_medium && (
            <img
              src={book.book_cover_large}
              alt={book.book_title}
              width={100}
              style={{ borderRadius: 4, objectFit: "contain" }}
            />
          )}
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: 4,
              justifyContent: "center",
            }}
          >
            <strong>{book.book_title}</strong>
            <span>{book.book_author}</span>
            {book.rating > 0 && (
              <span style={{ color: "#f59e0b" }}>
                {"★".repeat(book.rating)}
                {"☆".repeat(5 - book.rating)}
              </span>
            )}
            {book.review_text && <span>{book.review_text}</span>}
          </div>
        </a>
      ))}
    </div>
  );
}`;

export async function ApiDemoWithCode() {
  const highlighted = await highlight(CODE, {
    lang: "jsx",
    themes: { light: "github-light", dark: "github-dark" },
  });

  return (
    <ApiDemo
      highlightedCode={
        <CodeBlock>
          <pre>{highlighted}</pre>
        </CodeBlock>
      }
    />
  );
}
