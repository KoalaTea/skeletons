
import './App.css'
import { ExamplePage } from './pages/example';
import { createBrowserRouter, RouterProvider } from "react-router";

import { AuthorizationContextProvider } from './context/AuthorizationContext';


const router = createBrowserRouter([
  {
    path: "/",
    element: <ExamplePage />,
  }
])

function App() {
  return (
    <AuthorizationContextProvider>
      <RouterProvider router={router} />
    </AuthorizationContextProvider>
  )
}

export default App
