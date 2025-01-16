import React, { useState, useEffect } from 'react';
import { AgGridReact } from 'ag-grid-react';
import 'ag-grid-community/styles/ag-grid.css';
import 'ag-grid-community/styles/ag-theme-alpine.css';

function AgGridComponent() {
  const [rowData, setRowData] = useState([]);

  const columnDefs = [
    { field: 'Book_id' },
    { field: 'Title' },
    { field: 'Author' },
    { field: 'Summary' },
    { field: 'First_published' },
    { field: 'Last_updated' }
  ];

  useEffect(() => {
    fetch('http://localhost:8080/api/data')
      .then(response => response.json())
      .then(data => setRowData(data));
  }, []);

  return (
    <div className="ag-theme-alpine" style={{height: 400, width: 600}}>
      <AgGridReact
        columnDefs={columnDefs}
        rowData={rowData}
      />
    </div>
  );
}

export default AgGridComponent;