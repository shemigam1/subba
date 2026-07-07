"use client";

import { motion } from "motion/react";

interface SubbaLogoProps {
  className?: string;
  size?: "sm" | "md" | "lg";
  showText?: boolean;
}

export function SubbaLogo({ className = "", size = "lg", showText = true }: SubbaLogoProps) {
  const dimensions = {
    sm: { container: "w-8 h-8", svg: 24, dot: "w-1 h-1", gap: "gap-2", text: "text-xl", offset: -3 },
    md: { container: "w-16 h-16", svg: 50, dot: "w-2 h-2", gap: "gap-4", text: "text-2xl", offset: -5 },
    lg: { container: "w-32 h-32", svg: 100, dot: "w-4 h-4", gap: "gap-6", text: "text-4xl", offset: -10 },
  }[size];

  return (
    <div className={`flex items-center justify-center ${dimensions.gap} ${size === "lg" ? "flex-col" : "flex-row"} ${className}`}>
      {/* Animated Logo Graphic */}
      <div className={`relative ${dimensions.container} flex items-center justify-center`}>
        {/* The fading triangle */}
        <motion.svg
          width={dimensions.svg}
          height={dimensions.svg}
          viewBox="0 0 100 100"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          animate={{
            rotate: [0, 180, 360],
            scale: [1, 1.05, 1],
          }}
          transition={{
            duration: 20,
            repeat: Infinity,
            ease: "linear",
          }}
          className="absolute"
        >
          {/* A triangle with fading (blurred) edges */}
          <path
            d="M50 10L90 85H10L50 10Z"
            fill="url(#triangleGradient)"
            filter="url(#blur)"
            opacity="0.6"
          />
          <defs>
            <linearGradient id="triangleGradient" x1="50" y1="10" x2="50" y2="85" gradientUnits="userSpaceOnUse">
              <stop stopColor="#4f46e5" />
              <stop offset="1" stopColor="#4f46e5" stopOpacity="0" />
            </linearGradient>
            <filter id="blur" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="8" />
            </filter>
          </defs>
        </motion.svg>

        {/* The Three Dots */}
        {[0, 1, 2].map((i) => {
          const positions = [
            { top: "5%", left: "50%", transform: "translate(-50%, 0)" },
            { bottom: "10%", left: "15%" },
            { bottom: "10%", right: "15%" },
          ];
          return (
            <motion.div
              key={i}
              className={`absolute ${dimensions.dot} bg-brand-600 rounded-full`}
              style={positions[i]}
              animate={{
                y: [0, dimensions.offset, 0],
                scale: [1, 1.2, 1],
                opacity: [0.8, 1, 0.8],
              }}
              transition={{
                duration: 3,
                repeat: Infinity,
                ease: "easeInOut",
                delay: i * 0.4,
              }}
            />
          );
        })}
      </div>

      {/* Animated Brand Name */}
      {showText && (
        <motion.h1
          className={`${dimensions.text} font-extrabold tracking-tight text-slate-900`}
          animate={{
            opacity: [0.8, 1, 0.8],
          }}
          transition={{
            duration: 4,
            repeat: Infinity,
            ease: "easeInOut",
          }}
        >
          subba<span className="text-brand-600">.</span>
        </motion.h1>
      )}
    </div>
  );
}
