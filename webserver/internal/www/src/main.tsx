import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { InMemoryCache, ApolloClient, ApolloProvider } from '@apollo/client';

const REACT_APP_API_ENDPOINT = import.meta.env.VITE_APP_API_ENDPOINT;
console.log(REACT_APP_API_ENDPOINT)
console.log(import.meta.env.MODE)

const client = new ApolloClient({
  uri: `${REACT_APP_API_ENDPOINT}/graphql`,
  cache: new InMemoryCache(),
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ApolloProvider client={client}>
      <App />
    </ApolloProvider>
  </StrictMode>,
)
