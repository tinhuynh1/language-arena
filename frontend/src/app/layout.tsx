import type { Metadata } from "next";
import { AuthProvider } from "@/hooks/useAuth";
import Header from "@/components/layout/Header";
import "@/styles/globals.css";

export const metadata: Metadata = {
  title: "Lingo Sniper — Train Your Reflexes, Master Languages",
  description: "A real-time multiplayer vocabulary aim trainer. Combine CSGO-style reflex training with foreign language learning. Challenge friends in 1v1 duels!",
  keywords: ["vocabulary", "aim trainer", "language learning", "multiplayer", "csgo", "reflex"],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;700&family=Rajdhani:wght@500;600;700&display=swap" rel="stylesheet" />
      </head>
      <body>
        <AuthProvider>
          <Header />
          <main className="pt-[4.5rem] min-h-screen">
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  );
}
