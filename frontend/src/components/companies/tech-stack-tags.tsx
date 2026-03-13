interface TechStackTagsProps {
  tags: string[];
  limit?: number;
}

export function TechStackTags({ tags, limit = 5 }: TechStackTagsProps) {
  const visible = limit > 0 ? tags.slice(0, limit) : tags;
  const remaining = tags.length - visible.length;

  return (
    <div className="flex flex-wrap gap-1.5">
      {visible.map((tag) => (
        <span
          key={tag}
          className="inline-flex items-center rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-medium text-blue-700"
        >
          {tag}
        </span>
      ))}
      {remaining > 0 && (
        <span className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs text-gray-500">
          +{remaining}
        </span>
      )}
    </div>
  );
}
