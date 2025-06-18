import { redirect } from "react-router";
import type { Route } from "./+types/edit";

export async function loader({ params }: Route.LoaderArgs) {
  // This will be implemented when we build the nginx templates functionality
  return {
    template: {
      id: params.id,
      name: "Sample Template",
      content: "# This feature is not implemented yet"
    }
  };
}

export async function action({ request, params }: Route.ActionArgs) {
  // This will be implemented when we build the nginx templates functionality
  throw new Error("Not implemented yet");
}

export default function EditNginxTemplate({ loaderData }: Route.ComponentProps) {
  const { template } = loaderData || { template: { id: '', name: 'Unknown', content: '' } };

  return (
    <div className="container mx-auto py-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-2xl font-bold mb-6">Edit Nginx Template: {template.name}</h1>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-blue-700">
            This feature is coming soon. Template editing functionality will be implemented in Phase 4.
          </p>
        </div>
      </div>
    </div>
  );
}
