import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import {App} from './App.tsx';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {Route, Switch} from 'wouter';

const queryClient = new QueryClient();

createRoot(document.getElementById('root')!).render(
	<StrictMode>
		<QueryClientProvider client={queryClient}>
			<Switch>
				<Route path="/join/:lobbyId">
					{({ lobbyId }) => <App lobbyId={lobbyId} /> }
				</Route>
				<Route>
					<App />
				</Route>
			</Switch>
		</QueryClientProvider>
	</StrictMode>,
);
