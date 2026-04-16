import {
  HTMLAttributes,
  ThHTMLAttributes,
  TdHTMLAttributes,
  TableHTMLAttributes,
} from "react";
import { cn } from "../../lib/utils";

export const Table = ({
  className,
  ...props
}: TableHTMLAttributes<HTMLTableElement>) => {
  return (
    <div className="ui-table-wrap">
      <table className={cn("ui-table", className)} {...props} />
    </div>
  );
};

export const TableHeader = ({
  className,
  ...props
}: HTMLAttributes<HTMLTableSectionElement>) => {
  return <thead className={cn("ui-table__header", className)} {...props} />;
};

export const TableBody = ({
  className,
  ...props
}: HTMLAttributes<HTMLTableSectionElement>) => {
  return <tbody className={cn("ui-table__body", className)} {...props} />;
};

export const TableRow = ({
  className,
  ...props
}: HTMLAttributes<HTMLTableRowElement>) => {
  return <tr className={cn("ui-table__row", className)} {...props} />;
};

export const TableHead = ({
  className,
  ...props
}: ThHTMLAttributes<HTMLTableCellElement>) => {
  return <th className={cn("ui-table__head", className)} {...props} />;
};

export const TableCell = ({
  className,
  ...props
}: TdHTMLAttributes<HTMLTableCellElement>) => {
  return <td className={cn("ui-table__cell", className)} {...props} />;
};
