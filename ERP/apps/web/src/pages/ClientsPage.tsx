import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";
import { useAuth } from "../lib/auth";

interface ClientProfileDTO {
  phone: string;
  customer_found: boolean;
  balance: number;
  free_drinks_left: number;
  coffee_punches: number;
}

interface MenuProductDTO {
  id: string;
  name: string;
  category_name: string;
  base_price: number;
  sale_price: number | null;
}

export default function ClientsPage() {
  const { user } = useAuth();
  const [phone, setPhone] = useState("");
  const [submittedPhone, setSubmittedPhone] = useState("");

  const locationId = user?.location_id ?? "";
  const canQueryMenu = Boolean(locationId);

  const { data: profile, isFetching: profileLoading } = useQuery<ClientProfileDTO>({
    queryKey: ["client-profile", submittedPhone],
    queryFn: () =>
      api.get("/clients/profile", { params: { phone: submittedPhone } }).then((r) => r.data.data),
    enabled: submittedPhone.length >= 4,
    retry: false,
  });

  const { data: menu = [] } = useQuery<MenuProductDTO[]>({
    queryKey: ["client-menu", locationId],
    queryFn: () =>
      api.get("/menu", { params: { location_id: locationId } }).then((r) => r.data.data ?? []),
    enabled: canQueryMenu,
  });

  const [showLinkModal, setShowLinkModal] = useState(false);

  const deliveryLink = useMemo(
    () => (locationId ? `${window.location.origin}/order?location=${locationId}` : ""),
    [locationId]
  );

  const summaryTitle = useMemo(() => {
    if (!submittedPhone) return "Введите номер телефона";
    if (profile?.customer_found) return "Бонусный баланс";
    return "У клиента пока нет бонусов";
  }, [profile?.customer_found, submittedPhone]);

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <div className="mx-auto flex max-w-6xl flex-col gap-6 p-6 lg:p-8">
        <div className="rounded-2xl border border-gray-800 bg-gray-900/80 p-6 shadow-xl">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
            <div>
              <p className="text-sm font-medium uppercase tracking-[0.25em] text-emerald-400">
                Клиентский кабинет
              </p>
              <h1 className="mt-2 text-3xl font-bold">Проверьте баланс и откройте меню</h1>
              <p className="mt-2 max-w-2xl text-sm text-gray-400">
                Введите номер телефона клиента — система покажет его бонусный баланс, а также меню
                доступных позиций.
              </p>
            </div>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                setSubmittedPhone(phone.trim());
              }}
              className="flex w-full max-w-md gap-2"
            >
              <input
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+7 700 000 00 00"
                className="w-full rounded-xl border border-gray-700 bg-gray-800 px-4 py-3 text-sm text-white outline-none ring-0"
              />
              <button
                type="submit"
                className="rounded-xl bg-emerald-600 px-4 py-3 text-sm font-semibold text-white transition hover:bg-emerald-500"
              >
                Проверить
              </button>
            </form>
          </div>
        </div>

        {deliveryLink && (
          <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-5">
            <p className="text-sm font-medium text-emerald-300">Ссылка для доставки (QR)</p>
            <p className="mt-1 text-xs text-gray-400">
              Разместите QR с этой ссылкой в офисах — клиенты откроют меню и оформят доставку.
            </p>
            <div className="mt-3 flex flex-col gap-2 sm:flex-row">
              <input
                readOnly
                value={deliveryLink}
                onFocus={(e) => e.currentTarget.select()}
                className="w-full rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-200 outline-none"
              />
              <button
                onClick={() => {
                  navigator.clipboard?.writeText(deliveryLink);
                  alert("Ссылка скопирована!");
                }}
                className="shrink-0 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-emerald-500"
              >
                Копировать
              </button>
              <button
                onClick={() => setShowLinkModal(true)}
                className="shrink-0 rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-500"
              >
                Ссылка
              </button>
            </div>
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <section className="rounded-2xl border border-gray-800 bg-gray-900/80 p-6 shadow-xl">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-400">Баланс клиента</p>
                <h2 className="mt-1 text-xl font-semibold">{summaryTitle}</h2>
              </div>
              {profileLoading && <span className="text-sm text-emerald-400">Загрузка...</span>}
            </div>

            {submittedPhone && profile ? (
              <div className="mt-6 grid gap-4 md:grid-cols-2">
                <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 p-4">
                  <p className="text-sm text-emerald-300">Доступно бонусов</p>
                  <p className="mt-2 text-3xl font-bold text-emerald-400">
                    {profile.balance.toLocaleString("ru-RU")} ₸
                  </p>
                </div>
                <div className="rounded-xl border border-sky-500/20 bg-sky-500/10 p-4">
                  <p className="text-sm text-sky-300">Бесплатных кофе</p>
                  <p className="mt-2 text-3xl font-bold text-sky-400">{profile.free_drinks_left}</p>
                </div>
                <div className="rounded-xl border border-amber-500/20 bg-amber-500/10 p-4 md:col-span-2">
                  <p className="text-sm text-amber-300">Серийность кофе</p>
                  <p className="mt-2 text-2xl font-semibold text-amber-400">
                    {profile.coffee_punches} шт.
                  </p>
                </div>
              </div>
            ) : (
              <div className="mt-6 rounded-xl border border-dashed border-gray-700 bg-gray-800/50 p-6 text-sm text-gray-400">
                После ввода номера телефона здесь появится бонусный баланс клиента.
              </div>
            )}
          </section>

          <section className="rounded-2xl border border-gray-800 bg-gray-900/80 p-6 shadow-xl">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-400">Меню</p>
                <h2 className="mt-1 text-xl font-semibold">Доступные позиции</h2>
              </div>
            </div>

            {!canQueryMenu ? (
              <div className="mt-6 rounded-xl border border-dashed border-gray-700 bg-gray-800/50 p-6 text-sm text-gray-400">
                Чтобы отобразить меню, войдите в систему с привязкой к заведению.
              </div>
            ) : menu.length === 0 ? (
              <div className="mt-6 rounded-xl border border-dashed border-gray-700 bg-gray-800/50 p-6 text-sm text-gray-400">
                Меню пока пусто.
              </div>
            ) : (
              <div className="mt-6 space-y-2">
                {menu.map((product) => (
                  <div
                    key={product.id}
                    className="rounded-xl border border-gray-800 bg-gray-800/70 p-3"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="font-medium text-white">{product.name}</p>
                        <p className="text-sm text-gray-400">{product.category_name}</p>
                      </div>
                      <div className="text-right">
                        {product.sale_price != null ? (
                          <>
                            <p className="text-xs text-gray-500 line-through">
                              {product.base_price.toLocaleString("ru-RU")} ₸
                            </p>
                            <p className="font-semibold text-amber-400">
                              {product.sale_price.toLocaleString("ru-RU")} ₸
                            </p>
                          </>
                        ) : (
                          <p className="font-semibold text-emerald-400">
                            {product.base_price.toLocaleString("ru-RU")} ₸
                          </p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      </div>

      {/* Link display modal */}
      {showLinkModal && deliveryLink && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm rounded-2xl bg-white p-8 shadow-xl">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-bold text-gray-900">Ссылка для доставки</h2>
              <button
                onClick={() => setShowLinkModal(false)}
                className="text-2xl leading-none text-gray-400 hover:text-gray-600"
              >
                ×
              </button>
            </div>

            <div className="space-y-6">
              <div className="space-y-3">
                <p className="text-sm text-gray-700 font-medium">Ссылка для доставки:</p>
                <div className="bg-gray-100 p-3 rounded-lg border border-gray-300">
                  <p className="font-mono text-xs text-gray-900 break-all">{deliveryLink}</p>
                </div>
              </div>

              <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-900">
                <p className="font-semibold mb-2">📱 Используйте эту ссылку:</p>
                <ul className="text-xs space-y-1 ml-4 list-disc">
                  <li>Нажмите «Сгенерировать QR» чтобы создать код</li>
                  <li>Или скопируйте ссылку в любой QR-генератор</li>
                  <li>Распечатайте QR и разместите в офисах</li>
                </ul>
              </div>

              <div className="flex flex-col gap-2">
                <button
                  onClick={() => {
                    navigator.clipboard?.writeText(deliveryLink);
                    alert("Ссылка скопирована!");
                  }}
                  className="w-full rounded-lg bg-emerald-600 py-2.5 font-semibold text-white hover:bg-emerald-500"
                >
                  Копировать ссылку
                </button>
                <button
                  onClick={() => {
                    const qrUrl = `https://qr-server.com/api/qrcode?format=png&size=300x300&data=${encodeURIComponent(deliveryLink)}`;
                    window.open(qrUrl, "qr_generator", "width=400,height=500");
                  }}
                  className="w-full rounded-lg bg-amber-600 py-2.5 font-semibold text-white hover:bg-amber-500"
                >
                  🔗 Сгенерировать QR
                </button>
                <button
                  onClick={() => window.print()}
                  className="w-full rounded-lg bg-sky-600 py-2.5 font-semibold text-white hover:bg-sky-500"
                >
                  Печать
                </button>
                <button
                  onClick={() => setShowLinkModal(false)}
                  className="w-full rounded-lg bg-gray-200 py-2.5 font-semibold text-gray-900 hover:bg-gray-300"
                >
                  Закрыть
                </button>
              </div>

              <p className="text-xs text-gray-500 text-center">
                💡 Распечатайте этот экран или скопируйте ссылку и разместите в офисах
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
