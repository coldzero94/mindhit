"use client";

import { Check, X } from "lucide-react";
import type { SubscriptionPlan } from "@/api/generated/types.gen";

interface PlanComparisonTableProps {
  plans: SubscriptionPlan[];
  currentPlanId?: string;
}

export function PlanComparisonTable({
  plans,
  currentPlanId,
}: PlanComparisonTableProps) {
  const sortedPlans = [...plans].sort((a, b) => {
    const order = ["free", "pro", "enterprise"];
    return order.indexOf(a.id) - order.indexOf(b.id);
  });

  const features = [
    {
      name: "월간 토큰",
      getValue: (plan: SubscriptionPlan) =>
        plan.token_limit ? plan.token_limit.toLocaleString() : "무제한",
    },
    {
      name: "세션 보관",
      getValue: (plan: SubscriptionPlan) =>
        plan.session_retention_days ? `${plan.session_retention_days}일` : "무제한",
    },
    {
      name: "동시 세션",
      getValue: (plan: SubscriptionPlan) =>
        plan.max_concurrent_sessions
          ? `${plan.max_concurrent_sessions}개`
          : "무제한",
    },
    {
      name: "PNG 내보내기",
      getValue: (plan: SubscriptionPlan) => plan.features?.export_png ?? false,
      isBoolean: true,
    },
    {
      name: "SVG 내보내기",
      getValue: (plan: SubscriptionPlan) => plan.features?.export_svg ?? false,
      isBoolean: true,
    },
    {
      name: "Markdown 내보내기",
      getValue: (plan: SubscriptionPlan) => plan.features?.export_md ?? false,
      isBoolean: true,
    },
    {
      name: "JSON 내보내기",
      getValue: (plan: SubscriptionPlan) => plan.features?.export_json ?? false,
      isBoolean: true,
    },
    {
      name: "우선 지원",
      getValue: (plan: SubscriptionPlan) =>
        plan.features?.priority_support ?? false,
      isBoolean: true,
    },
    {
      name: "API 접근",
      getValue: (plan: SubscriptionPlan) => plan.features?.api_access ?? false,
      isBoolean: true,
    },
    {
      name: "팀 공유",
      getValue: (plan: SubscriptionPlan) =>
        plan.features?.team_sharing ?? false,
      isBoolean: true,
    },
    {
      name: "SSO 로그인",
      getValue: (plan: SubscriptionPlan) => plan.features?.sso ?? false,
      isBoolean: true,
    },
  ];

  return (
    <div className="overflow-x-auto">
      <table className="w-full border-collapse">
        <thead>
          <tr>
            <th className="text-left py-4 px-4 border-b border-gray-200 bg-gray-50 text-gray-600 font-medium">
              기능
            </th>
            {sortedPlans.map((plan) => (
              <th
                key={plan.id}
                className={`py-4 px-4 border-b border-gray-200 text-center font-semibold ${
                  currentPlanId === plan.id
                    ? "bg-blue-50 text-blue-700"
                    : "bg-gray-50 text-gray-900"
                }`}
              >
                {plan.name}
                {currentPlanId === plan.id && (
                  <span className="block text-xs font-normal text-blue-600 mt-1">
                    현재 플랜
                  </span>
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {features.map((feature, index) => (
            <tr
              key={feature.name}
              className={index % 2 === 0 ? "bg-white" : "bg-gray-50"}
            >
              <td className="py-3 px-4 border-b border-gray-100 text-gray-700">
                {feature.name}
              </td>
              {sortedPlans.map((plan) => {
                const value = feature.getValue(plan);
                return (
                  <td
                    key={plan.id}
                    className={`py-3 px-4 border-b border-gray-100 text-center ${
                      currentPlanId === plan.id ? "bg-blue-50/50" : ""
                    }`}
                  >
                    {feature.isBoolean ? (
                      value ? (
                        <Check className="w-5 h-5 text-green-500 mx-auto" />
                      ) : (
                        <X className="w-5 h-5 text-gray-300 mx-auto" />
                      )
                    ) : (
                      <span className="text-gray-900 font-medium">
                        {value as string}
                      </span>
                    )}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
