export type Agent = {
  id: string;
  name: string;
  role: string;
  trait: string;
  color: string;
  image: string;
  speaking_image?: string;
  expressions?: Record<string, string>;
  popular?: boolean;
  greeting?: string;
  robot?: boolean;
  skin?: string;
  hair?: string;
};

export async function loadWorkforce(): Promise<Agent[]> {
  const res = await fetch('/api/workforce');
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to load workforce');
  return data.agents ?? [];
}