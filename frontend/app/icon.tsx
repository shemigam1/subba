import { ImageResponse } from "next/og";

export const runtime = "edge";
export const size = {
  width: 32,
  height: 32,
};
export const contentType = "image/png";

export default function Icon() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
        }}
      >
        {/* The triangle */}
        <svg
          width="32"
          height="32"
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          style={{ position: "absolute" }}
        >
          <path
            d="M50 10L90 85H10L50 10Z"
            fill="#4f46e5"
            opacity="0.8"
          />
        </svg>
        {/* The 3 dots */}
        <div style={{ position: "absolute", top: "2%", left: "50%", transform: "translate(-50%, 0)", width: "6px", height: "6px", backgroundColor: "#4f46e5", borderRadius: "50%" }} />
        <div style={{ position: "absolute", bottom: "8%", left: "12%", width: "6px", height: "6px", backgroundColor: "#4f46e5", borderRadius: "50%" }} />
        <div style={{ position: "absolute", bottom: "8%", right: "12%", width: "6px", height: "6px", backgroundColor: "#4f46e5", borderRadius: "50%" }} />
      </div>
    ),
    {
      ...size,
    }
  );
}
