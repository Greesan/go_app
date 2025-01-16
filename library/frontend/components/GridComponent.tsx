// components/GridComponent.tsx
"use client";

import { AgGridReact } from 'ag-grid-react';
import { useEffect, useState, useCallback } from "react";
import type { ColDef, GridApi } from "ag-grid-community";
import { AllCommunityModule, ModuleRegistry } from "ag-grid-community";

ModuleRegistry.registerModules([AllCommunityModule]);

const columnDefs: ColDef[] = [
    { field: "Book_id", headerName: "ID" },
    { field: "Title" },
    { field: "Summary" },
    { field: "Author" },
    { field: "First_published", headerName: "First Published" },
    { field: "Last_updated", headerName: "Last Updated" },
  ];

const GridComponent = () => {
  const [rowData, setRowData] = useState<any[]>([]);
  const [gridApi, setGridApi] = useState<GridApi | null>(null);
  const fetchBooks = useCallback(() => {
    fetch("http://localhost:8080/api/books")
      .then((result) => result.json())
      .then((books) => {
        if (gridApi) {
          setRowData(books);
        }
      })
      .catch((error) => console.error('Error fetching books:', error));
  }, [gridApi]);

  useEffect(() => {
    fetchBooks();
    const intervalId = setInterval(fetchBooks, 5000);
    return () => clearInterval(intervalId);
  }, [fetchBooks]);
  const onGridReady = (params: { api: GridApi }) => {
    setGridApi(params.api);
  };

  return (
    <div style={{ width: "100%", height: "100vh" }}>
      <button onClick={fetchBooks}>Refresh Data</button>
      <AgGridReact
        rowData={rowData}
        columnDefs={columnDefs}
        pagination={true}
        paginationPageSize={10}
        onGridReady={onGridReady}
      />
    </div>
  );
};

export default GridComponent;