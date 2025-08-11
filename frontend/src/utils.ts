export function throwIfNotOk(fetchCall: Promise<Response>) {
	return fetchCall
		.then((response) => {
			return response.text()
				.then(text => {
					if (!response.ok) {
						throw new Error(text);
					}

					return text;
				});
		});
}

export const api = (str: string) => import.meta.env.VITE_SERVER_URL ?? `http://localhost:8080${str}`;