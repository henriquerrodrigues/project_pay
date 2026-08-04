import { NextResponse, type NextRequest } from "next/server";
import { AUTH_COOKIE } from "@/lib/api";

const PUBLIC_PATHS = ["/login", "/signup", "/api/login", "/api/signup"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  const isPublic = PUBLIC_PATHS.some(
    (p) => pathname === p || pathname.startsWith(`${p}/`)
  );
  if (isPublic) {
    return NextResponse.next();
  }

  const hasToken = Boolean(req.cookies.get(AUTH_COOKIE)?.value);
  if (!hasToken) {
    return NextResponse.redirect(new URL("/login", req.url));
  }

  return NextResponse.next();
}

// tudo menos assets do Next
export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
