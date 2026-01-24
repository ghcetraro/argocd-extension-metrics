//Wrapper for fetch
export const apiCall = (url: string, headers: Record<string, string>) => {
  return fetch(url, { headers })
    .then((res) => res.json())
    .then((res) => res)
    .catch((err) => {
      throw err;
    });
};

// This function is making an api call directly to the metrics server (bypassing ArgoCD proxy)
export function getDashBoard({
  applicationName,
  applicationNamespace,
  resourceType,
  project,
}: {
  applicationName: string;
  applicationNamespace: string;
  resourceType: string;
  project: string;
}) {
  // Direct connection to metrics server instead of using ArgoCD proxy
  const url = `http://argocd-metrics-dev.devops.ligo.live/api/applications/${applicationName}/groupkinds/${resourceType.toLowerCase()}/dashboards`;
  return fetch(url, {
    headers: getHeaders({ applicationName, applicationNamespace, project }),
  })
    .then((response) => response)
    .catch((err) => {
      throw err;
    });
}

//Creates and returns the custom headers needed for the argocd extensions.
export function getHeaders({
  applicationName,
  applicationNamespace,
  project,
}: {
  applicationName: string;
  applicationNamespace: string;
  project: string;
}) {
  const argocdApplicationName = `${applicationNamespace}:${applicationName}`;
  return {
    "Argocd-Application-Name": `${argocdApplicationName}`,
    "Argocd-Project-Name": `${project}`,
  };
}