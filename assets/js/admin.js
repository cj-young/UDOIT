import React from 'react'
import { createRoot } from 'react-dom/client'
import AdminApp from './Components/Admin/AdminApp'
import getInitialData from './getInitialData'
import Api from './Services/Api'

const root = createRoot(document.getElementById('root'))

getInitialData('api/admin/settings').then((data) => {
  const api = new Api(data.settings);
  root.render(<AdminApp api={api} {...data} />)
})
