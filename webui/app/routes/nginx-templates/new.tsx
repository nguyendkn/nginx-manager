import { redirect } from "react-router";
import type { Route } from "./+types/new";

export async function action({ request }: Route.ActionArgs) {
  // This will be implemented when we build the nginx templates functionality
  throw new Error("Not implemented yet");
}

export default function NewNginxTemplate() {
  return (
    <div className="container mx-auto py-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-2xl font-bold mb-6">Create New Nginx Template</h1>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-blue-700">
            This feature is coming soon. Template creation functionality will be implemented in Phase 4.
          </p>
        </div>
      </div>
    </div>
  );
}
