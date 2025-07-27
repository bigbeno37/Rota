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