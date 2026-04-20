import "./globals.css";
import { AuthProvider } from "../context/AuthContext";
import { CartProvider } from "../context/CartContext";
import { TopProgressProvider } from "../context/TopProgressContext";
import Header from "@/components/layout/Header";
import Footer from "@/components/layout/Footer";

export const metadata = {
  title: "Go E-commerce App",
  description: "Frontend for Go E-commerce Backend",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body suppressHydrationWarning={true}>
        <TopProgressProvider>
          <AuthProvider>
            <CartProvider>
              <div style={{
                display: 'flex',
                flexDirection: 'column',
                minHeight: '100vh'
              }}>
                <Header />
                <main style={{ flex: 1 }}>{children}</main>
                <Footer />
              </div>
            </CartProvider>
          </AuthProvider>
        </TopProgressProvider>
      </body>
    </html>
  );
}
