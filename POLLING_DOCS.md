# Real-Time Metrics Polling for QEMU VMs

This feature implements automatic real-time polling of virtual machine metrics using React Query and ORPC.

## Overview

The polling system automatically fetches VM metrics at configurable intervals, keeping the UI up-to-date with the latest VM status information.

## Implementation

### Backend
- ORPC API endpoints for VM metrics
- Lightweight HTTP-based polling (no WebSocket required initially)
- Support for both list and individual VM polling

### Frontend
- Custom `usePollingVMs` hook for easy integration
- Configurable polling intervals
- Automatic cache management with React Query
- Optional polling enable/disable

## Usage

### Basic Usage - List All VMs with Polling

```typescript
import { usePollingVMs } from '@/hooks';

export function VMsList() {
  // Polls every 5 seconds by default
  const { data: vms, isLoading, error } = usePollingVMs();
  
  return (
    <div>
      {vms?.map(vm => (
        <div key={vm.uuid}>{vm.name} - {vm.state}</div>
      ))}
    </div>
  );
}
```

### Custom Polling Interval

```typescript
// Poll every 2 seconds
const { data: vms } = usePollingVMs({ interval: 2000 });

// Poll every 10 seconds
const { data: vms } = usePollingVMs({ interval: 10000 });
```

### Conditional Polling

```typescript
// Enable/disable polling based on visibility
const { data: vms } = usePollingVMs({ 
  enabled: document.visibilityState === 'visible',
  interval: 5000 
});
```

### Individual VM Polling

```typescript
import { usePollingVM } from '@/hooks';

export function VMDetails({ vmUUID }) {
  const { data: vm, isLoading } = usePollingVM(vmUUID, { interval: 3000 });
  
  return (
    <div>
      <h3>{vm?.name}</h3>
      <p>CPU Time: {vm?.cpu_time_ns}</p>
      <p>Memory: {vm?.memory_kb} KB</p>
    </div>
  );
}
```

## Configuration

### Polling Intervals
- **Recommended**: 5000ms (5 seconds) - Good balance between freshness and server load
- **Fast**: 2000ms (2 seconds) - More frequent updates, higher server load
- **Slow**: 10000ms (10 seconds) - Less server load, older data

### Cache Settings
The cache is configured as:
- `staleTime`: Half the polling interval (data is considered fresh until then)
- `refetchInterval`: The configured polling interval
- `refetchOnWindowFocus`: Enabled (data updates when user focuses window)

## API Endpoints

### List VMs with Info
- **Endpoint**: `GET /qemu/virtual-machines/info`
- **Permission**: `RBAC_QEMU_READ`
- **Returns**: Array of `VirtualMachineWithInfo`
- **Polling**: Recommended every 5-10 seconds

### Get Single VM Info
- **Endpoint**: `GET /qemu/virtual-machines/:uuid/info`
- **Permission**: `RBAC_QEMU_READ`
- **Returns**: Single `VirtualMachineWithInfo`
- **Polling**: Recommended every 3-5 seconds

## Performance Considerations

1. **Stale Time**: Data is considered stale after half the polling interval, ensuring data freshness
2. **Query Deduplication**: React Query automatically deduplicates requests within the same moment
3. **Background Refetch**: Polling continues even if component is not visible (can be controlled with `enabled` option)
4. **Window Focus**: By default, polling refetches when user focuses the window again

## Best Practices

1. **Use Appropriate Intervals**
   - Use 5000ms for list views (less critical)
   - Use 3000ms for detail views (more critical)
   - Use 2000ms for real-time dashboards only

2. **Enable/Disable Wisely**
   ```typescript
   const [isVisible, setIsVisible] = useState(true);
   
   useEffect(() => {
     const handleVisibility = () => {
       setIsVisible(document.visibilityState === 'visible');
     };
     document.addEventListener('visibilitychange', handleVisibility);
     return () => document.removeEventListener('visibilitychange', handleVisibility);
   }, []);
   
   const { data } = usePollingVMs({ enabled: isVisible });
   ```

3. **Handle Errors Gracefully**
   ```typescript
   const { data, error, isLoading } = usePollingVMs();
   
   if (error && !data) return <ErrorState />;
   if (isLoading && !data) return <LoadingState />;
   ```

## Files Modified

- `frontend/src/hooks/use-polling-vms.ts` - New polling hook implementation
- `frontend/src/hooks/index.ts` - Export polling hook
- `frontend/src/routes/app.vms.tsx` - Updated to use polling hook
- `internal/services/qemu_service.go` - Added CreateVirtualMachine handler
- `internal/server/routes.go` - Added create VM route
- `frontend/src/lib/routers/qemu.ts` - Added createVirtualMachine endpoint
- `frontend/src/routes/qemu/create-vm-dialog.tsx` - New VM creation dialog

## Future Enhancements

1. **WebSocket Support**: For truly real-time updates
2. **Server-Sent Events (SSE)**: Push notifications from server
3. **Adaptive Polling**: Automatically adjust interval based on data change rate
4. **Pause on Inactivity**: Stop polling when page is inactive for extended period
5. **Exponential Backoff**: Reduce polling frequency on repeated errors

## Testing

To test the polling functionality:

1. Open the VMs page
2. Open browser DevTools Network tab
3. Verify API calls are made every 5 seconds
4. Update VM state externally
5. Verify UI updates reflect new state within polling interval

## Troubleshooting

### VMs not updating?
- Check browser DevTools Network tab for API calls
- Verify RBAC_QEMU_READ permission is granted
- Check if `enabled` option is set to false

### Server overloaded?
- Increase polling interval (use 10000ms or more)
- Disable polling on list view, only enable on detail view
- Implement conditional polling based on visibility

### Stale data?
- Reduce polling interval
- Manually invalidate queries: `queryClient.invalidateQueries()`
- Check network conditions in DevTools
