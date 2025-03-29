import type { Metadata } from "next";
import { JetBrains_Mono } from "next/font/google";
import Script from "next/script"; // 1. Import Script
import "./globals.css";

const jetBrainsMono = JetBrains_Mono({
  variable: "--font-jetbrains-mono",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "Project Starbyte",
  description: "Embark on an interstellar adventure in your terminal!",
  icons: {
    icon: "/favicon.ico",
    shortcut: "/favicon.ico",
    apple: "/favicon.ico",
  },
  openGraph: {
    title: "Project Starbyte",
    description: "Embark on an interstellar adventure in your terminal!",
    url: "https://projectstarbyte.com",
    siteName: "Project Starbyte",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
      },
    ],
    locale: "en-US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Project Starbyte",
    description: "Embark on an interstellar adventure in your terminal!",
    images: ["/og-image.png"],
  },
  themeColor: "#000000",
  appleWebApp: {
    capable: true,
    statusBarStyle: "black-translucent",
  },
  manifest: "/manifest.json",
  viewport: "width=device-width, initial-scale=1",
  robots: {
    index: true,
    follow: true,
  },
  verification: {
    google: "google-site-verification=your-google-verification-code",
  },
};

const GTM_ID = "GTM-MFLC76LK";

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      {/* 2. Add the GTM <script> using next/script */}
      <Script id="google-tag-manager" strategy="afterInteractive">
        {`
          (function(w,d,s,l,i){w[l]=w[l]||[];w[l].push({'gtm.start':
          new Date().getTime(),event:'gtm.js'});var f=d.getElementsByTagName(s)[0],
          j=d.createElement(s),dl=l!='dataLayer'?'&l='+l:'';j.async=true;j.src=
          'https://www.googletagmanager.com/gtm.js?id='+i+dl;f.parentNode.insertBefore(j,f);
          })(window,document,'script','dataLayer','${GTM_ID}');
        `}
      </Script>

      <body className={`${jetBrainsMono.variable} antialiased`}>
        {/* 3. Add the GTM <noscript> tag */}
        <noscript
          dangerouslySetInnerHTML={{
            __html: `<iframe src="https://www.googletagmanager.com/ns.html?id=${GTM_ID}"
                      height="0" width="0" style="display:none;visibility:hidden"></iframe>`,
          }}
        />
        {children}
      </body>
    </html>
  );
}