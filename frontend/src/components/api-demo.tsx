"use client";

import { useState, type ReactNode } from "react";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "fumadocs-ui/components/tabs";
import { CodeBlock } from "fumadocs-ui/components/codeblock";

interface Book {
  book_title: string;
  book_author: string;
  book_cover_large: string;
  book_cover_medium: string;
  book_link: string;
  avg_rating: number;
  rating?: number;
  review_text?: string;
  review_published_on?: string;
}

interface ShelfResponse {
  count: number;
  books: Book[];
}

function parseUserId(input: string): string | null {
  const trimmed = input.trim();
  if (/^\d+$/.test(trimmed)) return trimmed;
  const match = trimmed.match(/goodreads\.com\/user\/show\/(\d+)/);
  return match ? match[1] : null;
}

function Stars({ rating }: { rating: number }) {
  return (
    <span className="text-yellow-500">
      {"★".repeat(rating)}
      {"☆".repeat(5 - rating)}
    </span>
  );
}

export function ApiDemo({ highlightedCode }: { highlightedCode?: ReactNode }) {
  const [input, setInput] = useState("");
  const [books, setBooks] = useState<Book[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit() {
    const userId = parseUserId(input);
    if (!userId) {
      setError(
        "Could not parse a user ID. Paste a Goodreads profile URL or enter an ID.",
      );
      return;
    }

    setLoading(true);
    setError(null);
    setBooks(null);

    const url = `https://api.piratereads.com/${userId}/read?per_page=5`;

    try {
      const res = await fetch(url);
      if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
      const data: ShelfResponse = await res.json();
      console.log("piratereads response:", data);
      setBooks(data.books);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "Something went wrong";
      setError(msg);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Tabs defaultValue="demo" className="not-prose">
      <TabsList>
        <TabsTrigger value="demo">Demo</TabsTrigger>
        <TabsTrigger value="code">Code</TabsTrigger>
      </TabsList>

      <TabsContent value="demo">
        <div className="flex flex-col gap-4">
          <div className="flex gap-2">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
              placeholder="https://www.goodreads.com/user/show/12345-name"
              className="flex-1 rounded-lg border bg-fd-secondary px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-fd-primary"
            />
            <button
              type="button"
              onClick={handleSubmit}
              disabled={loading || !input.trim()}
              className="rounded-lg bg-fd-primary px-4 py-2 text-sm font-medium text-fd-primary-foreground disabled:opacity-50"
            >
              {loading ? "Loading..." : "Get read books"}
            </button>
          </div>

          {error && (
            <p className="rounded-lg border border-red-500/20 bg-red-500/10 p-3 text-sm text-red-500">
              {error}
            </p>
          )}

          {books && books.length === 0 && (
            <p className="text-sm text-fd-muted-foreground">
              No books found for this user.
            </p>
          )}

          {books && books.length > 0 && (
            <div className="grid gap-3">
              {books.map((book) => (
                <a
                  key={book.book_link}
                  href={book.book_link}
                  target="_blank"
                  rel="noreferrer"
                  className="flex gap-4 rounded-lg border p-3 hover:bg-fd-secondary transition-colors"
                >
                  {book.book_cover_medium && (
                    <img
                      src={book.book_cover_medium}
                      alt={book.book_title}
                      className="w-12 h-18 rounded object-cover shrink-0"
                    />
                  )}
                  <div className="flex flex-col gap-1 min-w-0">
                    <span className="font-medium text-sm">
                      {book.book_title}
                    </span>
                    <span className="text-xs text-fd-muted-foreground">
                      {book.book_author}
                    </span>
                    {book.rating != null && book.rating > 0 && (
                      <Stars rating={book.rating} />
                    )}
                    {book.review_text && (
                      <p className="text-xs text-fd-muted-foreground line-clamp-2">
                        {book.review_text}
                      </p>
                    )}
                  </div>
                </a>
              ))}
            </div>
          )}
        </div>
      </TabsContent>

      <TabsContent value="code">
        {highlightedCode ?? (
          <CodeBlock>
            <pre>
              <code>Loading...</code>
            </pre>
          </CodeBlock>
        )}
      </TabsContent>
    </Tabs>
  );
}
