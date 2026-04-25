import type { Metadata } from "next";
import { DM_Sans, Inter, JetBrains_Mono } from "next/font/google";
import { AuthProvider } from "@/hooks/useAuth";
import { LocaleProvider } from "@/i18n/LocaleProvider";
import Header from "@/components/layout/Header";
import "@/styles/globals.css";

const dmSans = DM_Sans({
  subsets: ["latin"],
  variable: "--font-dm-sans",
  weight: ["400", "500", "600", "700"],
});

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  weight: ["400", "500", "600", "700"],
});

const jetBrainsMono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-jetbrains",
  weight: ["400", "500", "700"],
});

export const metadata: Metadata = {
  title: "LinguaLeap — Master Vocabulary, One Word at a Time",
  description: "An interactive vocabulary learning platform. Practice English and Chinese words through engaging quiz sessions. Study solo or challenge friends in real-time.",
  keywords: ["vocabulary", "language learning", "education", "study", "quiz", "English", "Chinese", "HSK", "CEFR"],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" translate="no" className={`${dmSans.variable} ${inter.variable} ${jetBrainsMono.variable}`}>
      <head>
        <meta name="google" content="notranslate" />
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
