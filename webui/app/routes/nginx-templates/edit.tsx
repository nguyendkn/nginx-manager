import { useParams } from "react-router";
import { useState, useEffect } from "react";

interface Template {
  id: string;
  name: string;
  content: string;
}

export default function EditNginxTemplate() {
  const { id } = useParams();
  const [template, setTemplate] = useState<Template>({
    id: id || '',
    name: 'Unknown',
    content: ''
  });

  useEffect(() => {
    // This will be implemented when we build the nginx templates functionality
    // For now, just set sample data
    if (id) {
      setTemplate({
        id,
        name: "Sample Template",
        content: "# This feature is not implemented yet"
      });
    }
  }, [id]);

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
