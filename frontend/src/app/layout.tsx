import type { Metadata } from "next";
import { AuthProvider } from "@/hooks/useAuth";
import { LocaleProvider } from "@/i18n/LocaleProvider";
import Header from "@/components/layout/Header";
import "@/styles/globals.css";

export const metadata: Metadata = {
  title: "LinguaLeap — Master Vocabulary, One Word at a Time",
  description: "An interactive vocabulary learning platform. Practice English and Chinese words through engaging quiz sessions. Study solo or challenge friends in real-time.",
  keywords: ["vocabulary", "language learning", "education", "study", "quiz", "English", "Chinese", "HSK", "CEFR"],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" translate="no">
      <head>
        <meta name="google" content="notranslate" />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet" />
      </head>
      <body>
        <AuthProvider>
          <LocaleProvider>
            <Header />
            <main className="pt-[4.5rem] min-h-screen">
              {children}
            </main>
          </LocaleProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
