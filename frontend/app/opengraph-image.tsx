import { ImageResponse } from "next/og";

export const runtime = "edge";
export const alt = "Subba | Smarter subscriptions for African businesses";
export const size = {
  width: 1200,
  height: 630,
};
export const contentType = "image/png";

export default async function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#4f46e5", // brand-600
          color: "white",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "20px" }}>
          {/* Logo representation */}
          <div
            style={{
              width: "60px",
              height: "60px",
              backgroundColor: "white",
              borderRadius: "12px",
            }}
          />
          <span style={{ fontSize: "72px", fontWeight: "bold", letterSpacing: "-0.05em" }}>
            Subba
          </span>
        </div>
        <p style={{ marginTop: "40px", fontSize: "36px", opacity: 0.9, textAlign: "center", maxWidth: "800px", lineHeight: 1.4 }}>
          Recurring billing that never drops a payment. Built on Nomba.
        </p>
      </div>
    ),
    {
      ...size,
    }
  );
}
