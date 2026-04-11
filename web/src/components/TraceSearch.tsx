import { useState } from "react";

interface Props {
  onSearch: (requestId: string) => void;
  loading: boolean;
}

export function TraceSearch({ onSearch, loading }: Props) {
  const [value, setValue] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = value.trim();
    if (trimmed) {
      onSearch(trimmed);
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ display: "flex", gap: 8 }}>
      <input
        type="text"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        placeholder="Enter request_id..."
        style={{
          flex: 1,
          padding: "8px 12px",
          border: "1px solid #d1d5db",
          borderRadius: 6,
          fontSize: 14,
          outline: "none",
        }}
      />
      <button
        type="submit"
        disabled={loading || !value.trim()}
        style={{
          padding: "8px 20px",
          background: loading ? "#9ca3af" : "#2563eb",
          color: "#fff",
          border: "none",
          borderRadius: 6,
          fontSize: 14,
          cursor: loading ? "not-allowed" : "pointer",
        }}
      >
        {loading ? "Loading..." : "Fetch"}
      </button>
    </form>
  );
}
