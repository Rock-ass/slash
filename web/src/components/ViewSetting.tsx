import { Button, Divider, Option, Select } from "@mui/joy";
import { toast } from "react-hot-toast";
import useViewStore from "../stores/v1/view";
import Icon from "./Icon";
import Dropdown from "./common/Dropdown";

const ViewSetting = () => {
  const viewStore = useViewStore();
  const order = viewStore.getOrder();
  const { field, direction } = order;
  const displayStyle = viewStore.displayStyle || "full";

  const handleReset = () => {
    viewStore.setOrder({ field: "name", direction: "asc" });
    toast.success("Order reset");
  };

  return (
    <Dropdown
      trigger={
        <button>
          <Icon.Settings2 className="w-4 h-auto text-gray-500" />
        </button>
      }
      actionsClassName="!mt-3 !-right-2"
      actions={
        <div className="w-52 p-2 pt-0 gap-2 flex flex-col justify-start items-start" onClick={(e) => e.stopPropagation()}>
          <div className="w-full flex flex-row justify-between items-center mt-1">
            <span className="text-sm font-medium">View order</span>
            <Button size="sm" variant="plain" color="neutral" onClick={handleReset}>
              <Icon.RefreshCw className="w-4 h-auto text-gray-500" />
            </Button>
          </div>
          <div className="w-full flex flex-row justify-between items-center">
            <span className="text-sm shrink-0 mr-2">Order by</span>
            <Select size="sm" value={field} onChange={(_, value) => viewStore.setOrder({ field: value as any })}>
              <Option value={"name"}>Name</Option>
              <Option value={"updatedTs"}>CreatedAt</Option>
              <Option value={"createdTs"}>UpdatedAt</Option>
              <Option value={"view"}>Visits</Option>
            </Select>
          </div>
          <div className="w-full flex flex-row justify-between items-center">
            <span className="text-sm shrink-0 mr-2">Direction</span>
            <Select size="sm" value={direction} onChange={(_, value) => viewStore.setOrder({ direction: value as any })}>
              <Option value={"asc"}>ASC</Option>
              <Option value={"desc"}>DESC</Option>
            </Select>
          </div>
          <Divider />
          <div className="w-full flex flex-row justify-between items-center">
            <span className="text-sm shrink-0 mr-2">Display</span>
            <Select size="sm" value={displayStyle} onChange={(_, value) => viewStore.setDisplayStyle(value as any)}>
              <Option value={"full"}>Full</Option>
              <Option value={"compact"}>Compact</Option>
            </Select>
          </div>
        </div>
      }
    ></Dropdown>
  );
};

export default ViewSetting;
