"use client";

import { motion } from "motion/react";

export function SubbaLogo({ className = "" }: { className?: string }) {
  return (
    <div className={`flex flex-col items-center justify-center gap-6 ${className}`}>
      {/* Animated Logo Graphic */}
      <div className="relative w-32 h-32 flex items-center justify-center">
        {/* The fading triangle */}
        <motion.svg
          width="100"
          height="100"
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
          // Positions corresponding to a triangle
          const positions = [
            { top: "5%", left: "50%" },
            { bottom: "10%", left: "15%" },
            { bottom: "10%", right: "15%" },
          ];
          return (
            <motion.div
              key={i}
              className="absolute w-4 h-4 bg-brand-600 rounded-full"
              style={positions[i]}
              animate={{
                y: [0, -10, 0],
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
      <motion.h1
        className="text-4xl font-extrabold tracking-tight text-slate-900"
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
    </div>
  );
}
