import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './Components/App'
import getInitialData from './getInitialData'
import { api } from './Services/Api'

const root = createRoot(document.getElementById('root'))

getInitialData('api/settings').then((data) => {
  api.setSettings(data.settings);
  root.render(<App { ...data } />);
})
