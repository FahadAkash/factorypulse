using System;
using System.Security.Cryptography;
using System.Text;

namespace FactoryPulseSim.Utils
{
    /// <summary>
    /// Generates deterministic UUIDv5 (SHA-1 based) from a namespace and name.
    /// Compliant with RFC 4122 for deterministic event_id deduplication.
    /// </summary>
    public static class UuidV5
    {
        // Standard DNS namespace UUID: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
        public static readonly Guid NamespaceDns = new Guid("6ba7b810-9dad-11d1-80b4-00c04fd430c8");

        public static Guid Create(string name)
        {
            return Create(NamespaceDns, name);
        }

        public static Guid Create(Guid namespaceId, string name)
        {
            if (string.IsNullOrEmpty(name))
            {
                throw new ArgumentNullException(nameof(name));
            }

            byte[] namespaceBytes = namespaceId.ToByteArray();
            SwapByteOrder(namespaceBytes);

            byte[] nameBytes = Encoding.UTF8.GetBytes(name);
            byte[] combinedBytes = new byte[namespaceBytes.Length + nameBytes.Length];
            Buffer.BlockCopy(namespaceBytes, 0, combinedBytes, 0, namespaceBytes.Length);
            Buffer.BlockCopy(nameBytes, 0, combinedBytes, namespaceBytes.Length, nameBytes.Length);

            byte[] hash;
            using (var sha1 = SHA1.Create())
            {
                hash = sha1.ComputeHash(combinedBytes);
            }

            byte[] newGuid = new byte[16];
            Buffer.BlockCopy(hash, 0, newGuid, 0, 16);

            // Set version 5 (bits 4-7 of time_hi_and_version)
            newGuid[6] = (byte)((newGuid[6] & 0x0F) | (5 << 4));
            // Set variant to RFC 4122 (bits 6-7 of clock_seq_hi_and_reserved)
            newGuid[8] = (byte)((newGuid[8] & 0x3F) | 0x80);

            SwapByteOrder(newGuid);
            return new Guid(newGuid);
        }

        private static void SwapByteOrder(byte[] guidBytes)
        {
            Swap(guidBytes, 0, 3);
            Swap(guidBytes, 1, 2);
            Swap(guidBytes, 4, 5);
            Swap(guidBytes, 6, 7);
        }

        private static void Swap(byte[] b, int i, int j)
        {
            byte t = b[i];
            b[i] = b[j];
            b[j] = t;
        }
    }
}

