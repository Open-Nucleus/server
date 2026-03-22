import { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface JsonViewerProps {
  data: unknown;
  collapsed?: boolean;
}

export function JsonViewer({ data, collapsed = false }: JsonViewerProps) {
  const [isCollapsed, setIsCollapsed] = useState(collapsed);

  const formatted = JSON.stringify(data, null, 2);

  return (
    <div
      className={cn(
        "rounded-[var(--radius-md)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "bg-[var(--color-paper)] dark:bg-[var(--color-paper-dark)]",
        "overflow-hidden",
      )}
    >
      {/* Toggle header */}
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className={cn(
          "flex items-center gap-1.5 w-full px-3 py-1.5 text-left cursor-pointer",
          "text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)]",
          "hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]",
          "transition-colors duration-100",
        )}
      >
        {isCollapsed ? <ChevronRight size={12} /> : <ChevronDown size={12} />}
        JSON
      </button>

      {!isCollapsed && (
        <pre
          className={cn(
            "px-3 pb-3 pt-0 overflow-x-auto text-[13px] leading-relaxed",
            "font-mono whitespace-pre",
          )}
        >
          <code>
            <Colorized json={formatted} />
          </code>
        </pre>
      )}
    </div>
  );
}

/* ---------- syntax coloring ---------- */

function Colorized({ json }: { json: string }) {
  // Tokenize JSON string into colored spans
  const tokens = tokenize(json);

  return (
    <>
      {tokens.map((token, i) => (
        <span key={i} className={token.className}>
          {token.text}
        </span>
      ))}
    </>
  );
}

interface Token {
  text: string;
  className: string;
}

function tokenize(json: string): Token[] {
  const tokens: Token[] = [];
  // Match JSON tokens: strings, numbers, booleans, null, structural chars
  const regex =
    /("(?:\\.|[^"\\])*")\s*:|("(?:\\.|[^"\\])*")|(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)|(\btrue\b|\bfalse\b)|(\bnull\b)|([{}[\]:,])|(\s+)/g;

  let match: RegExpExecArray | null;
  let lastIndex = 0;

  while ((match = regex.exec(json)) !== null) {
    // Capture any text before the match (shouldn't happen with well-formed JSON)
    if (match.index > lastIndex) {
      tokens.push({
        text: json.slice(lastIndex, match.index),
        className: "",
      });
    }

    if (match[1] !== undefined) {
      // Key (string followed by colon)
      tokens.push({
        text: match[1],
        className: "text-[var(--color-muted)]",
      });
      // The colon and whitespace are captured as part of the lookahead
      const colonAndSpace = json.slice(
        match.index + match[1].length,
        match.index + match[0].length,
      );
      tokens.push({ text: colonAndSpace, className: "" });
    } else if (match[2] !== undefined) {
      // String value
      tokens.push({
        text: match[2],
        className: "text-[var(--color-success)]",
      });
    } else if (match[3] !== undefined) {
      // Number
      tokens.push({
        text: match[3],
        className: "text-[var(--color-info)]",
      });
    } else if (match[4] !== undefined) {
      // Boolean
      tokens.push({
        text: match[4],
        className: "text-[#9C27B0] dark:text-[#CE93D8]",
      });
    } else if (match[5] !== undefined) {
      // Null
      tokens.push({
        text: match[5],
        className: "text-[var(--color-error)]",
      });
    } else if (match[6] !== undefined) {
      // Structural characters
      tokens.push({
        text: match[6],
        className: "text-[var(--color-muted)] opacity-60",
      });
    } else if (match[7] !== undefined) {
      // Whitespace
      tokens.push({ text: match[7], className: "" });
    }

    lastIndex = match.index + match[0].length;
  }

  // Remaining text
  if (lastIndex < json.length) {
    tokens.push({ text: json.slice(lastIndex), className: "" });
  }

  return tokens;
}
