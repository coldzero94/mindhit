import { GoogleSignInButton } from "./GoogleSignInButton";

export function LoginPrompt() {
  return (
    <div className="p-4">
      <div className="text-center mb-4">
        <h1 className="text-lg font-bold text-gray-900">MindHit</h1>
        <p className="text-xs text-gray-500 mt-1">
          Sign in with Google to get started
        </p>
      </div>

      <GoogleSignInButton />

      <p className="text-center text-xs text-gray-500 mt-4">
        By signing in, you agree to our Terms of Service and Privacy Policy.
      </p>
    </div>
  );
}
